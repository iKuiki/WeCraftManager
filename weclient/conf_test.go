package weclient

import (
	"github.com/google/uuid"
	"github.com/ikuiki/go-component/language"
	"github.com/ikuiki/storer"
	"math/rand"
	"os"
	"testing"
)

func TestConfSaveAndLoad(t *testing.T) {
	var c config
	c.UserName = uuid.New().String()
	count := rand.Intn(100)
	for i := 0; i < count; i++ {
		c.McChatrooms = append(c.McChatrooms, uuid.New().String())
	}
	filepath := uuid.New().String() + "_test.json"
	storer := storer.MustNewFileStorer(filepath)
	c.storer = storer
	defer os.Remove(filepath)
	err := c.Save()
	if err != nil {
		t.Fatalf("save config fail: %+v", err)
	}
	var c2 config
	c2.storer = storer
	err = c2.Load()
	if err != nil {
		t.Fatalf("load config fail: %+v", err)
	}
	if c.UserName != c2.UserName {
		t.Fatalf("c.UserName diff with c2.UserName\nc: %v\nc2: %v", c.UserName, c2.UserName)
	}
	if len(language.ArrayDiff(c.McChatrooms, c2.McChatrooms).([]string)) != len(language.ArrayDiff(c2.McChatrooms, c.McChatrooms).([]string)) {
		t.Fatalf("c.McChatrooms diff with c2.McChatrooms\nc: %v\nc2: %v", c.McChatrooms, c2.McChatrooms)
	}
}
