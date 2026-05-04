package dictionary

type Dictionary interface {
	Words() []string //words returns all known words
	Add(word string) //inserting a new word to the dictionary
	Load() error
	Save() error
}

// we wil add this lateron allowing multiple streams Like file,db,crwl etc..
type WordSource interface {
	LoadWords() ([]string, error)
}
