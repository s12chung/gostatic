/*
Package goodreads gets Goodreads books and ratings through the Goodsreads API.
*/
package goodreads

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/s12chung/gostatic/go/lib/utils"
)

const booksCacheFilename = "books.json"

// RatingMap Counts the ratings and makes a map of ratingNumber=>count
func RatingMap(books []*Book) map[int]int {
	ratingMap := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	i := 0
	for _, book := range books {
		ratingMap[book.Rating]++
		i++
	}
	return ratingMap
}

type Client struct {
	Settings *Settings
	log      logrus.FieldLogger
}

func NewClient(settings *Settings, log logrus.FieldLogger) *Client {
	return &Client{
		settings,
		log,
	}
}

// GetBooks returns read Books with their review ratings from the GoodsApi.
//
// It caches these results in goodreads.Settings.CachePath for later.
func (client *Client) GetBooks() ([]*Book, error) {
	err := utils.MkdirAll(client.Settings.CachePath)
	if err != nil {
		return nil, err
	}

	bookMap, err := client.getBooks(client.Settings.UserID)
	if err != nil {
		return nil, err
	}
	return toBooks(bookMap), nil
}

func toBooks(bookMap map[string]*Book) []*Book {
	books := make([]*Book, len(bookMap))
	i := 0
	for _, book := range bookMap {
		books[i] = book
		i++
	}
	return books
}

func (client *Client) getBooks(userID int) (map[string]*Book, error) {
	bookMap := client.readBooksCache()
	if client.Settings.invalid() {
		client.log.Warn("Invalid Goodreads Settings, skipping Goodreads API calls")
		return bookMap, nil
	}
	bookMap = client.GetBooksRequest(userID, bookMap)
	return bookMap, client.saveBooksCache(bookMap)
}

type jsonBooksRoot struct {
	Books map[string]*Book `json:"books"`
}

func (client *Client) readBooksCache() map[string]*Book {
	jsonRoot := jsonBooksRoot{}
	booksCachePath := path.Join(client.Settings.CachePath, booksCacheFilename)

	_, err := os.Stat(booksCachePath)
	if os.IsNotExist(err) {
		client.log.Infof("%v does not exist - %v", booksCachePath, err)
		return nil
	}

	bytes, err := ioutil.ReadFile(booksCachePath) // #nosec G304
	if err != nil {
		client.log.Warnf("error reading %v - %v", booksCachePath, err)
		return nil
	}

	err = json.Unmarshal(bytes, &jsonRoot)
	if err != nil {
		client.log.Warnf("error reading %v - %v", booksCachePath, err)
		return nil
	}

	if jsonRoot.Books == nil {
		client.log.Warnf("key books in %v is nil", booksCachePath)
	}
	client.log.Infof("Loaded %v books from %v", len(jsonRoot.Books), booksCachePath)
	return jsonRoot.Books
}

func (client *Client) saveBooksCache(bookMap map[string]*Book) error {
	bytes, err := json.MarshalIndent(jsonBooksRoot{bookMap}, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(client.Settings.CachePath, booksCacheFilename), bytes, 0755)
}

type xmlBookResponse struct {
	XMLName  xml.Name   `xml:"GoodreadsResponse"`
	PageData xmlReviews `xml:"reviews"`
}

func (response *xmlBookResponse) HasMore() bool {
	return response.PageData.PageEnd < response.PageData.TotalBooks
}

type xmlReviews struct {
	XMLName xml.Name `xml:"reviews"`

	Books []*Book `xml:"review"`

	PageStart  int `xml:"start,attr"`
	PageEnd    int `xml:"end,attr"`
	TotalBooks int `xml:"total,attr"`
}

// GetBooksRequest requests for the books and goes through pagination for the given userID
func (client *Client) GetBooksRequest(userID int, bookMap map[string]*Book) map[string]*Book {
	if bookMap == nil {
		bookMap = map[string]*Book{}
	}

	initialLoad := len(bookMap) == 0
	if initialLoad {
		client.log.Info("Loading all data from goodreads API")
	}

	totalAPIBooks := 0
	booksAdded := 0

	defer func() {
		if len(bookMap) < totalAPIBooks {
			client.log.Warnf("bookMap has %v elements, while there are %v books in the API", len(bookMap), totalAPIBooks)
		} else {
			client.log.Infof("bookMap has all %v books", totalAPIBooks)
		}
	}()

	err := client.paginateGet(
		func(page int) (resp *http.Response, err error) {
			return client.requestGetBooks(userID, initialLoad, page)
		},
		func(bytes []byte) (bool, error) {
			bookResponse := xmlBookResponse{}
			err := xml.Unmarshal(bytes, &bookResponse)
			if err != nil {
				return false, err
			}

			totalAPIBooks = bookResponse.PageData.TotalBooks

			for _, book := range bookResponse.PageData.Books {
				if len(bookMap) >= totalAPIBooks {
					return false, nil
				}
				if _, contains := bookMap[book.ID]; !contains {
					booksAdded++
					client.log.Infof("%v. %v", booksAdded, book.ReviewString())
					book.convertDates()
					bookMap[book.ID] = book
				}
			}
			return bookResponse.HasMore() && len(bookMap) < totalAPIBooks, nil
		},
	)
	if err != nil {
		client.log.Warnf("paginateGet error - %v", err)
	}
	return bookMap
}

func (client *Client) requestGetBooks(userID int, initialLoad bool, page int) (resp *http.Response, err error) {
	perPage := client.Settings.PerPage
	if initialLoad {
		perPage = client.Settings.MaxPerPage
	}

	queryParams := map[string]string{
		"v":  "2",
		"id": strconv.Itoa(userID),

		"key": client.Settings.APIKey,

		"shelf": "read",

		"page":     strconv.Itoa(page),
		"per_page": strconv.Itoa(perPage),
		"sort":     "date_added",
		"order":    "d",
	}

	url := fmt.Sprintf("%v/review/list?%v", client.Settings.APIURL, utils.ToSimpleQuery(queryParams))
	client.log.Infof("GET %v", url)
	return http.Get(url) // #nosec G107
}

func (client *Client) paginateGet(request func(page int) (resp *http.Response, err error), callback func(bytes []byte) (bool, error)) error {
	rateLimit := time.Duration(client.Settings.RateLimit) * time.Millisecond
	ticker := time.NewTicker(rateLimit)
	defer ticker.Stop()

	page := 1
	hasMore := true
	for hasMore {
		response, err := request(page)
		if err != nil {
			return err
		}
		bytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		err = response.Body.Close()
		if err != nil {
			return err
		}

		hasMore, err = callback(bytes)
		if err != nil {
			return err
		}
		if hasMore {
			page++
			client.log.Infof("Sleeping for %v...", rateLimit)
			<-ticker.C
		}
	}
	return nil
}

type Date time.Time

func (date Date) Equal(u Date) bool {
	return time.Time(date).Equal(time.Time(u))
}

func (date *Date) UnmarshalXML(decoder *xml.Decoder, startElement xml.StartElement) error {
	var stringValue string

	err := decoder.DecodeElement(&stringValue, &startElement)
	if err != nil {
		return err
	}

	t, err := time.Parse(time.RubyDate, stringValue)
	if err != nil {
		return err
	}

	*date = Date(t)
	return nil
}

type Book struct {
	XMLName xml.Name `xml:"review" json:"-"`
	ID      string   `xml:"id" json:"id"`
	Title   string   `xml:"book>title" json:"title"`
	Authors []string `xml:"book>authors>author>name" json:"authors"`
	Isbn    string   `xml:"book>isbn" json:"isbn"`
	Isbn13  string   `xml:"book>isbn13" json:"isbn13"`
	Rating  int      `xml:"rating" json:"rating"`

	XMLDateAdded    Date      `xml:"date_added" json:"-"`
	XXMLDateUpdated Date      `xml:"date_updated" json:"-"`
	DateAdded       time.Time `xml:"-" json:"date_added"`
	DateUpdated     time.Time `xml:"-" json:"date_updated"`
}

func (book *Book) convertDates() {
	book.DateAdded = time.Time(book.XMLDateAdded)
	book.DateUpdated = time.Time(book.XXMLDateUpdated)
}

// ReviewString returns the String representation of a Book: "The Book Title" by Berry, Jerry & Daisy ****
func (book *Book) ReviewString() string {
	return fmt.Sprintf("\"%v\" by %v %v", book.Title, utils.SliceList(book.Authors), strings.Repeat("*", book.Rating))
}

// SortedDate returns the recommended date to Sort on
func (book *Book) SortedDate() time.Time {
	return book.DateAdded
}
