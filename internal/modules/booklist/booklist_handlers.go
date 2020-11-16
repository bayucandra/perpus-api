package booklist

import (
	"context"
	"github.com/BIQDev/perpus-api/internal/db"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type IMongoBooklist struct {
	ID    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	User  string             `json:"user" bson:"user"`
	Title string             `json:"title" bson:"title"`
	Image string             `json:"image" bson:"image"`
}

type booklistHandlers struct {
}

func (*booklistHandlers) Create(w http.ResponseWriter, r *http.Request) {
	var err error
	params := mux.Vars(r)

	if err = r.ParseMultipartForm(10 << 20); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	uploadedFile, handler, err := r.FormFile("image")
	if err != nil {
		log.Println("Error Retrieving the File")
		log.Println(err)
		return
	}
	defer uploadedFile.Close()

	username := params["username"]
	imgPathDir := "web/assets/" + username

	if _, err = os.Stat(imgPathDir); os.IsNotExist(err) {
		err = os.MkdirAll(imgPathDir, 0755)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	imgPathFile := imgPathDir + "/" + strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + handler.Filename
	targetFile, err := os.OpenFile(imgPathFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer targetFile.Close()
	if _, err := io.Copy(targetFile, uploadedFile); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	coll := db.Mongo.DB().Collection("booklist")
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
	res, err := coll.InsertOne(ctx, &IMongoBooklist{
		User: username ,
		Title: r.FormValue("title"),
		Image: imgPathFile,
	})

	log.Println(res, err)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

var BooklistHandlers = &booklistHandlers{}