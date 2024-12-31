package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


const connectionString  = "mongodb+srv://<your_DB_username>:<your_DB_password>@cluster0.ahx9n.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"

const dbName = "phoneBook"
const colName = "contacts"

var collection *mongo.Collection
var tmpl = template.Must(template.ParseGlob("templates/*"))

type Contact struct {
    ID    primitive.ObjectID `bson:"_id,omitempty"`
    Name  string             `bson:"name"`
    Phone string             `bson:"phone"`
}


//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// connect monogoDB
func initDB() {
    clientOptions := options.Client().ApplyURI(connectionString)

    client, err := mongo.Connect(context.TODO(), clientOptions)

    if err != nil {
        log.Fatal(err)
    }
	fmt.Println("Db connection successful")
    
    collection = client.Database(dbName).Collection(colName)

	fmt.Println("Collection instance is ready")
}


//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//Dispaly saved contact
func listContacts(w http.ResponseWriter, r *http.Request){
	cursor, err := collection.Find(context.TODO(),bson.M{})

	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(context.TODO())

	var contacts []Contact
    for cursor.Next(context.TODO()){
        var contact Contact
        if err := cursor.Decode(&contact); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        contacts = append(contacts, contact)
	}
    tmpl.ExecuteTemplate(w, "index.html", contacts)
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// Add a new contact
func addContact(w http.ResponseWriter, r *http.Request) {
    
    if r.Method == http.MethodPost {
        name := r.FormValue("name")
        phone := r.FormValue("phone")
        contact := Contact{Name: name, Phone: phone}

        _, err := collection.InsertOne(context.TODO(), contact)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        http.Redirect(w, r, "/", http.StatusSeeOther)
        // fmt.Println("contact saved")
    }
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//delete a contact
func deleteContact(w http.ResponseWriter, r *http.Request){

    r.ParseForm()
    id := r.FormValue("id")

    success := deleteContactFromDatabase(id)

    if success {
        http.Redirect(w,r,"/", http.StatusSeeOther)
        // fmt.Println("delete success")
    } else {
        http.Error(w, "Failed to delete contact",http.StatusInternalServerError)
    }
}

func deleteContactFromDatabase(id string) bool {
    client, err := mongo.Connect(context.TODO(),options.Client().ApplyURI(connectionString))

    if err != nil {
        log.Fatal(err)
    }

    defer client.Disconnect(context.TODO())

    collection := client.Database(dbName).Collection(colName)

    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        fmt.Println("Invalid ID format:", err)
        return false
    }

    filter := bson.M{"_id": objID}

    _, err = collection.DeleteOne(context.TODO(), filter)
    if err != nil {
        fmt.Println("Error deleting document:", err)
        return false
    }
    return true
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//updata contact
func updateContact(w http.ResponseWriter, r *http.Request){

    if r.Method == http.MethodGet {

        id := r.URL.Query().Get("id")
        
        objID, err := primitive.ObjectIDFromHex(id)
        if err != nil {
            log.Fatal(err)
        }

        var contact Contact
        collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&contact)

        tmpl.ExecuteTemplate(w, "update.html", contact)
    }

    if r.Method == http.MethodPost {
        id := r.FormValue("id")
        name := r.FormValue("name")
        phone := r.FormValue("phone")

        objID, err := primitive.ObjectIDFromHex(id)
        if err != nil {
           log.Fatal(err)
        }

        filter := bson.M{"_id": objID}
        update := bson.M{
            "$set": bson.M{
                "name":  name,
                "phone": phone,
            },
        }

        _, err = collection.UpdateOne(context.TODO(), filter, update)
        if err != nil {
            log.Fatal(err)
        }

        http.Redirect(w, r, "/", http.StatusSeeOther)
        // fmt.Println("Updated one contact")
    }
}


//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
func main() {
    initDB()

    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    http.HandleFunc("/", listContacts)
    http.HandleFunc("/add", addContact)
    http.HandleFunc("/delete",deleteContact)
    http.HandleFunc("/update",updateContact)

    fmt.Println("Server running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

