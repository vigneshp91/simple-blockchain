package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Book struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Author     string `json:"author"`
	ISBN       string `json:"isbn"`
	LaunchDate string `json:"launch_date"`
}

type BookTransaction struct {
	BookID       string `json:"id"`
	Buyer        string `json:"buyer"`
	PurchaseDate string `json:"purchase_date"`
	IsGenesis    bool   `json:"is_genesis"`
}

type BlockChain struct {
	blocks []*Block
}

type Block struct {
	Position  int
	TimeStamp string
	PrevHash  string
	Hash      string
	Data      BookTransaction
}

var blockChain *BlockChain
var books []*Book

func newBlockChain() *BlockChain {
	return &BlockChain{[]*Block{genesisBlock()}}
}

func genesisBlock() *Block {
	return generateBlock(&Block{}, BookTransaction{IsGenesis: true})
}

func main() {
	blockChain = newBlockChain()
	books = make([]*Book, 0)
	r := mux.NewRouter()
	r.HandleFunc("/block", getBlocks).Methods("GET")
	r.HandleFunc("/book", createBook).Methods("POST")
	r.HandleFunc("/block", createBlock).Methods("POST")
	r.HandleFunc("/books", getBooks).Methods("GET")

	log.Println("Listening on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}

func getBlocks(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(blockChain.blocks)
	return

}

func getBooks(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(books)
	return
}

func createBook(w http.ResponseWriter, r *http.Request) {
	var book *Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		log.Print("unable to parse json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to parse json"))
		return
	}
	data := fmt.Sprintf("%s,%s", book.Name, book.ISBN)
	book.ID = fmt.Sprintf("%x", md5.Sum([]byte(data)))
	books = append(books, book)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("book added successfully"))

}

func createBlock(w http.ResponseWriter, r *http.Request) {
	var bookTran BookTransaction
	if err := json.NewDecoder(r.Body).Decode(&bookTran); err != nil {
		log.Print("unable to parse json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to parse json"))
		return
	}
	if !isValidBook(bookTran.BookID) {
		log.Print("unable to parse json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to parse json"))
		return
	}
	prevBlock := blockChain.getLastBlock()
	block := generateBlock(prevBlock, bookTran)

	if !validBlock(prevBlock, block) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to create block"))
		return
	}
	blockChain.blocks = append(blockChain.blocks, block)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("block created"))
	return

}

func isValidBook(id string) bool {
	isValid := false
	for _, bid := range books {
		if id == bid.ID {
			isValid = true
		}
	}
	return isValid
}

func validBlock(prevBlock, block *Block) bool {
	if prevBlock.Hash != block.PrevHash {
		return false
	}
	return true
}

func (bc *BlockChain) getLastBlock() *Block {
	return bc.blocks[len(bc.blocks)-1]
}

func (b *Block) generateHash() {
	by, _ := json.Marshal(b.Data)
	data := []byte(string(rune(b.Position)) + string(by) + string(b.PrevHash) + b.TimeStamp)

	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))

}
func generateBlock(prevBlock *Block, bt BookTransaction) *Block {
	block := &Block{}
	block.PrevHash = prevBlock.Hash
	block.Data = bt
	block.Position = prevBlock.Position + 1
	block.TimeStamp = time.Now().String()
	block.generateHash()
	return block

}
