package main

import (
  "strings"
  "regexp"
)

type Store struct {
  Nodes map[rune]*Store
  isTerminal bool
}

type Link struct {
  Url string
  Depth int
}


var MainChannel  chan *Link = make(chan *Link)
var StoreChannel chan *string = make(chan *string)

var MainStore Store = Store{make(map[rune]*Store),false}

var ThreadCount int = 100

var MaxSearchDepth int = 3

var StartLink *Link = &Link{"http://en.wikipedia.org/wiki/Adolf_Hitler",0}

func main() {
  for i := 0;i < ThreadCount; i += 1 {
    go StartCrawler(HandleNewLink)
  }
  go StartStore()
  MainChannel<-StartLink
  return
}

// pulls links from the main channel
// the action takes in the current link, the title for that page, and its
func StartCrawler(Action func (*Link,string,string)) {
  for ;; {
    NextLink := <-MainChannel
    body := NextLink.UrlGet()
    title := TitleGet(body)
    Action(NextLink,title,body)
  }
}

func HandleNewLink(L *Link, title String, body String) {
  if L.Depth != MaxSearchDepth {
    Links = getLinks(body,L.Depth)
    for i,l := range Links {
      MainChannel<-l
      StoreChannel<-l.Url
    }
  }
}

// reads in from the store channel and adds to the store
func StartStore() {
  for ;; {
    s := <-StoreChannel
    s1 := &s
    MainStore.insert(s)
  }
}

// gets the html from the given link
func (self *Link) UrlGet() string {

}

/////////////////////////////////////////////
// utilities
/////////////////////////////////////////////

var RXC = regexp.MustCompile

var TitleRegexp       *Regexp = RXC("<title>.*<title>")
var MainDivHeadRegexp *Regexp = RXC("<!-- bodyContent --><div id=\"mw-content-text.*<!-- /bodyContent -->")
var LinkRegexp         *Regexp = RXC("<a href=\"/wiki/.*\".*>.*</a>")

// get the title from the wikipedia page
func TitleGet(body string) string {
  return TitleRegexp.FindString(body)
}

// get all Links from the content of the body of a page
func getLinks(body string, currentdepth int) []*Link {
  content := getContent(body)
  depth = currentdepth + 1
  links   := LinkRegex.FindAllString(content)
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
  self.insertSlice(strings.Split(name,""))
}

// helper for Store.insert(string)
func (self *Store) insertSlice(name []rune){
  if(len(name) == 0) {
    self.isTerminal = true
  } else {
    nextChar := name[0];
    nextStore := self.Nodes[nextChar]
    if nextStore == nil {
      nextStore = &Store{make(map[rune]Store, false}
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
    fmt.PrintLine(acc)
  }
  for c,s := self.Nodes {
    s.PrintString(acc + c)
  }
}
