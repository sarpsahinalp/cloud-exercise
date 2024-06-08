package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"slices"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Defines a "model" that we can use to communicate with the
// frontend or the database
type BookStore struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	BookName   string
	BookAuthor string
	BookISBN   string
	BookPages  int
	BookYear   int
}

type Book struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Author string `json:"author"`
	ISBN   string `json:"isbn"`
	Pages  int    `json:"pages"`
	Year   int    `json:"year"`
}

// Wraps the "Template" struct to associate a necessary method
// to determine the rendering procedure
type Template struct {
	tmpl *template.Template
}

// Preload the available templates for the view folder.
// This builds a local "database" of all available "blocks"
// to render upon request, i.e., replace the respective
// variable or expression.
// For more on templating, visit https://jinja.palletsprojects.com/en/3.0.x/templates/
// to get to know more about templating
// You can also read Golang's documentation on their templating
// https://pkg.go.dev/text/template
func loadTemplates() *Template {
	return &Template{
		tmpl: template.Must(template.ParseGlob("views/*.html")),
	}
}

// Method definition of the required "Render" to be passed for the Rendering
// engine.
// Contraire to method declaration, such syntax defines methods for a given
// struct. "Interfaces" and "structs" can have methods associated with it.
// The difference lies that interfaces declare methods whether struct only
// implement them, i.e., only define them. Such differentiation is important
// for a compiler to ensure types provide implementations of such methods.
func (t *Template) Render(w io.Writer, name string, data interface{}, ctx echo.Context) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}

// Here we make sure the connection to the database is correct and initial
// configurations exists. Otherwise, we create the proper database and collection
// we will store the data.
// To ensure correct management of the collection, we create a return a
// reference to the collection to always be used. Make sure if you create other
// files, that you pass the proper value to ensure communication with the
// database
// More on what bson means: https://www.mongodb.com/docs/drivers/go/current/fundamentals/bson/
func prepareDatabase(client *mongo.Client, dbName string, collecName string) (*mongo.Collection, error) {
	db := client.Database(dbName)

	names, err := db.ListCollectionNames(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, err
	}
	if !slices.Contains(names, collecName) {
		cmd := bson.D{{"create", collecName}}
		var result bson.M
		if err = db.RunCommand(context.TODO(), cmd).Decode(&result); err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	coll := db.Collection(collecName)
	return coll, nil
}

// Here we prepare some fictional data and we insert it into the database
// the first time we connect to it. Otherwise, we check if it already exists.
func prepareData(client *mongo.Client, coll *mongo.Collection) {
	startData := []BookStore{
		{
			BookName:   "The Vortex",
			BookAuthor: "José Eustasio Rivera",
			BookISBN:   "958-30-0804-4",
			BookPages:  292,
			BookYear:   1924,
		},
		{
			BookName:   "Frankenstein",
			BookAuthor: "Mary Shelley",
			BookISBN:   "978-3-649-64609-9",
			BookPages:  280,
			BookYear:   1818,
		},
		{
			BookName:   "The Black Cat",
			BookAuthor: "Edgar Allan Poe",
			BookISBN:   "978-3-99168-238-7",
			BookPages:  280,
			BookYear:   1843,
		},
	}

	// This syntax helps us iterate over arrays. It behaves similar to Python
	// However, range always returns a tuple: (idx, elem). You can ignore the idx
	// by using _.
	// In the topic of function returns: sadly, there is no standard on return types from function. Most functions
	// return a tuple with (res, err), but this is not granted. Some functions
	// might return a ret value that includes res and the err, others might have
	// an out parameter.
	for _, book := range startData {
		cursor, err := coll.Find(context.TODO(), book)
		var results []BookStore
		if err = cursor.All(context.TODO(), &results); err != nil {
			panic(err)
		}
		if len(results) > 1 {
			log.Fatal("more records were found")
		} else if len(results) == 0 {
			result, err := coll.InsertOne(context.TODO(), book)
			if err != nil {
				panic(err)
			} else {
				fmt.Printf("%+v\n", result)
			}

		} else {
			for _, res := range results {
				cursor.Decode(&res)
				fmt.Printf("%+v\n", res)
			}
		}
	}
}

// Generic method to perform "SELECT * FROM BOOKS" (if this was SQL, which
// it is not :D ), and then we convert it into an array of map. In Golang, you
// define a map by writing map[<key type>]<value type>{<key>:<value>}.
// interface{} is a special type in Golang, basically a wildcard...
func findAllBooks(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	var ret []map[string]interface{}
	for _, res := range results {
		ret = append(ret, map[string]interface{}{
			"ID":         res.ID.Hex(),
			"BookName":   res.BookName,
			"BookAuthor": res.BookAuthor,
			"BookISBN":   res.BookISBN,
			"BookPages":  res.BookPages,
		})
	}

	return ret
}

func getAllBooks(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	var ret []map[string]interface{}
	for _, res := range results {
		ret = append(ret, map[string]interface{}{
			"id":     res.ID.Hex(),
			"name":   res.BookName,
			"author": res.BookAuthor,
			"isbn":   res.BookISBN,
			"pages":  res.BookPages,
			"year":   res.BookYear,
		})
	}

	return ret
}

func findAllAuthors(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	var ret []map[string]interface{}
	for _, res := range results {
		ret = append(ret, map[string]interface{}{
			"ID":         res.ID.Hex(),
			"BookAuthor": res.BookAuthor,
		})
	}

	return ret
}

func findAllYears(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	var ret []map[string]interface{}
	for _, res := range results {
		ret = append(ret, map[string]interface{}{
			"ID":       res.ID.Hex(),
			"BookYear": res.BookYear,
		})
	}

	return ret
}

// Returns true if there is a duplicate in the database
func checkIfDuplicateExists(coll *mongo.Collection, book BookStore) bool {
	filter := bson.M{
		"bookname":   book.BookName,
		"bookauthor": book.BookAuthor,
		"bookisbn":   book.BookISBN,
		"bookpages":  book.BookPages,
		"bookyear":   book.BookYear,
	}

	// Perform the FindOne operation
	res := coll.FindOne(context.TODO(), filter)

	return res.Err() == nil
}

func saveBook(coll *mongo.Collection, newBook BookStore) []map[string]interface{} {
	res, err := coll.InsertOne(context.TODO(), newBook)
	if err != nil {
		return nil
	}

	var ret []map[string]interface{}
	ret = append(ret, map[string]interface{}{
		"ID": res.InsertedID,
	})

	return ret

}

func updateBook(coll *mongo.Collection, updatedBook BookStore) {
	filter := bson.M{
		"_id": updatedBook.ID,
	}

	update := bson.M{"$set": bson.M{
		"bookname":   updatedBook.BookName,
		"bookauthor": updatedBook.BookAuthor,
		"bookisbn":   updatedBook.BookISBN,
		"bookpages":  updatedBook.BookPages,
		"bookyear":   updatedBook.BookYear,
	}}

	_, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return
	}
}

func deleteBook(coll *mongo.Collection, id primitive.ObjectID) {
	filter := bson.M{
		"_id": id,
	}
	_, err := coll.DeleteOne(context.TODO(), filter)
	if err != nil {
		return
	}
}

func convertToBookstore(book Book) BookStore {
	var bookStore BookStore
	if book.ID != "" {
		bookStore.ID, _ = primitive.ObjectIDFromHex(book.ID)
	}
	bookStore.BookAuthor = book.Author
	bookStore.BookISBN = book.ISBN
	bookStore.BookName = book.Name
	bookStore.BookPages = book.Pages
	bookStore.BookYear = book.Year
	return bookStore
}

func main() {
	// Connect to the database. Such defer keywords are used once the local
	// context returns; for this case, the local context is the main function
	// By user defer function, we make sure we don't leave connections
	// dangling despite the program crashing. Isn't this nice? :D
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := os.Getenv("DATABASE_URI")
	if len(uri) == 0 {
		fmt.Printf("failure to load env variable\n")
		os.Exit(1)
	}
	// TODO: make sure to pass the proper username, password, and port
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))

	// This is another way to specify the call of a function. You can define inline
	// functions (or anonymous functions, similar to the behavior in Python)
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	// You can use such name for the database and collection, or come up with
	// one by yourself!
	coll, err := prepareDatabase(client, "exercise-1", "information")

	prepareData(client, coll)

	// Here we prepare the server
	e := echo.New()

	// Define our custom renderer
	e.Renderer = loadTemplates()

	// Log the requests. Please have a look at echo's documentation on more
	// middleware
	e.Use(middleware.Logger())

	e.Static("/css", "css")

	// Endpoint definition. Here, we divided into two groups: top-level routes
	// starting with /, which usually serve webpages. For our RESTful endpoints,
	// we prefix the route with /api to indicate more information or resources
	// are available under such route.
	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", nil)
	})

	e.GET("/books", func(c echo.Context) error {
		books := findAllBooks(coll)
		return c.Render(200, "book-table", books)
	})

	e.GET("/authors", func(c echo.Context) error {
		authors := findAllAuthors(coll)
		return c.Render(200, "author-table", authors)
	})

	e.GET("/years", func(c echo.Context) error {
		years := findAllYears(coll)
		return c.Render(200, "year-table", years)
	})

	e.GET("/search", func(c echo.Context) error {
		return c.Render(200, "search-bar", nil)
	})

	e.GET("/create", func(c echo.Context) error {
		return c.NoContent(304)
	})

	e.GET("/api/books", func(c echo.Context) error {
		books := getAllBooks(coll)
		return c.JSON(200, books)
	})

	e.POST("/api/books", func(c echo.Context) error {
		var book Book
		c.Bind(&book)
		toPost := convertToBookstore(book)
		if checkIfDuplicateExists(coll, toPost) {
			return c.JSON(304, "Duplicate not allowed")
		}
		res := saveBook(coll, toPost)
		return c.JSON(200, res)
	})

	e.PUT("/api/books", func(c echo.Context) error {
		var book Book
		c.Bind(&book)
		toUpdate := convertToBookstore(book)
		if checkIfDuplicateExists(coll, toUpdate) {
			return c.JSON(201, "Duplicate not allowed")
		}
		updateBook(coll, toUpdate)
		return c.JSON(200, "Updated the book")
	})

	e.DELETE("/api/books/:id", func(c echo.Context) error {
		id := c.Param("id")
		objectId, _ := primitive.ObjectIDFromHex(id)
		deleteBook(coll, objectId)
		return c.JSON(200, "Succesfully deleted entry")
	})

	e.Logger.Fatal(e.Start(":3030"))
}
