package main

import (
  "fmt"
  "strings"
  "regexp"
  "net/http"
  "io/ioutil"
)

type Store struct {
  Nodes map[string]*Store
  isTerminal bool
}

type Link struct {
  Url string
  Depth int
}


var WorkerChannel  chan *Link = make(chan *Link, 100)
var StoreChannel chan *string = make(chan *string, 100)
var ReplyChannel chan int     = make(chan int, 100)
var LinkChannel  chan *Link   = make(chan *Link, 10000)

var MainStore Store = Store{make(map[string]*Store),false}

var ThreadCount int = 100

var MaxSearchDepth int = 3

var StartLink *Link = &Link{"http://en.wikipedia.org/wiki/Adolf_Hitler",0}

func main() {
  fmt.Printf("lauching main worker threads")
  for i := 0;i < ThreadCount; i += 1 {
    go StartCrawler(HandleNewLink)
  }
  fmt.Printf("worker threads launched\n")
  fmt.Printf("launching store thread\n")
  go StartStore()
  fmt.Printf("store thread launched\n")
  fmt.Printf("starting main loop\n")
  WorkerChannel<-StartLink
  ResponceCount := 0
  MaxResponce := 1
  NextStep := 0
  for DepthCount := 0; DepthCount <= MaxSearchDepth; DepthCount += 1 {
    fmt.Printf("Starting Search level %d\n",DepthCount)
    for ;ResponceCount != MaxResponce; {
      fmt.Printf("%d\n",ResponceCount)
      nextInt := <-ReplyChannel
      ResponceCount += 1
      NextStep += nextInt
    }
    fmt.Printf("Search at depth %d completed\n",DepthCount)
    ResponceCount = 0
    MaxResponce = NextStep
    NextStep = 0
    l := <-LinkChannel
    suc := true
    for ;suc; {
      WorkerChannel<-l
      select {
        case l = <-LinkChannel:
        default: suc = false
      }
    }
  }
  MainStore.Print()
  return
}

// pulls links from the main channel
// the action takes in the current link, the title for that page, and its
func StartCrawler(Action func (*Link,string,string)) {
  for ;; {
    NextLink := <-WorkerChannel
    body := NextLink.UrlGet()
    title := TitleGet(body)
    Action(NextLink,title,body)
  }
}

func HandleNewLink(L *Link, title string, body string) {
  if L.Depth != MaxSearchDepth {
    Links := getLinks(body,L.Depth)
    for i,l := range Links {
      LinkChannel<-l
      StoreChannel<-&l.Url
      ReplyChannel<-i+1
    }
  }
}

// reads in from the store channel and adds to the store
func StartStore() {
  for ;; {
    s := <-StoreChannel
    s1 := *s
    MainStore.insert(s1)
    MainStore.Print()
  }
}

var httpClient *http.Client = &http.Client{}
// gets the html from the given link
func (self *Link) UrlGet() string {
  resp, _ := httpClient.Get(self.Url)
  body, _ := ioutil.ReadAll(resp.Body)
  resp.Body.Close()
  return string(body)
}

/////////////////////////////////////////////
// utilities
/////////////////////////////////////////////

var RXC = regexp.MustCompile

var TitleRegexp       *regexp.Regexp = RXC("<title>.*<title>")
var MainDivHeadRegexp *regexp.Regexp = RXC("<!-- bodyContent --><div id=\"mw-content-text.*<!-- /bodyContent -->")
var LinkRegexp         *regexp.Regexp = RXC("<a href=\"/wiki/.*\".*>.*</a>")

// get the title from the wikipedia page
func TitleGet(body string) string {
  return TitleRegexp.FindString(body)
}

// get all Links from the content of the body of a page
func getLinks(body string, currentdepth int) []*Link {
  content := getContent(body)
  depth := currentdepth + 1
  links   := LinkRegexp.FindAllString(content,100)
  var retLinks []*Link = make([]*Link,len(links))
  for i,s := range links {
    retLinks[i] = &Link{strings.SplitAfterN(s,"\"",3)[1],depth}
  }
  return retLinks
}

// get the content div from a wikipedia page
func getContent(body string) string {
  return MainDivHeadRegexp.FindString(body)
}



/////////////////////////////////////////////
// Store
/////////////////////////////////////////////

// add that string to the store.
func (self *Store) insert(name string){
  if(len(name) == 0) {
    self.isTerminal = true
  } else {
    nextChar := strings.SplitN(name,"",2)[0]
    nextStore := self.Nodes[nextChar]
    if nextStore == nil {
      nextStore = &Store{make(map[string]*Store), false}
      self.Nodes[nextChar] = nextStore
    }
    if len(name) == 1 {
      nextStore.isTerminal = true
    } else { 
      nextStore.insert(name[1:])
    }
  }
}

// prints out every string the store
func (self *Store) Print() {
  self.PrintString("")
}

// accumulatory helpter for store printer
func (self *Store) PrintString(acc string) {
  if self.isTerminal {
    fmt.Printf("%s\n",acc)
  }
  for c,s := range self.Nodes {
    s.PrintString(acc + c)
  }
}
