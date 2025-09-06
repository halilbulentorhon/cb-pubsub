package model

import (
	"testing"
	"time"
)

func TestCreatePubSubDoc(t *testing.T) {
	before := time.Now().Unix()
	doc := CreatePubSubDoc[string]()
	after := time.Now().Unix()

	if doc.CreationDate < before || doc.CreationDate > after {
		t.Errorf("CreationDate = %d, want between %d and %d", doc.CreationDate, before, after)
	}

	if doc.Messages == nil {
		t.Error("Messages is nil, want empty slice")
	}

	if len(doc.Messages) != 0 {
		t.Errorf("Messages length = %d, want 0", len(doc.Messages))
	}

	if cap(doc.Messages) != 0 {
		t.Errorf("Messages capacity = %d, want 0", cap(doc.Messages))
	}
}

func TestCreatePubSubDoc_DifferentTypes(t *testing.T) {
	intDoc := CreatePubSubDoc[int]()
	if intDoc.Messages == nil {
		t.Error("int Messages is nil")
	}
	if len(intDoc.Messages) != 0 {
		t.Errorf("int Messages length = %d, want 0", len(intDoc.Messages))
	}

	type customStruct struct {
		Name string
		ID   int
	}
	structDoc := CreatePubSubDoc[customStruct]()
	if structDoc.Messages == nil {
		t.Error("struct Messages is nil")
	}
	if len(structDoc.Messages) != 0 {
		t.Errorf("struct Messages length = %d, want 0", len(structDoc.Messages))
	}

	mapDoc := CreatePubSubDoc[map[string]interface{}]()
	if mapDoc.Messages == nil {
		t.Error("map Messages is nil")
	}
	if len(mapDoc.Messages) != 0 {
		t.Errorf("map Messages length = %d, want 0", len(mapDoc.Messages))
	}
}

func TestCreateEmptyMessages(t *testing.T) {
	stringMessages := CreateEmptyMessages[string]()
	if stringMessages == nil {
		t.Error("string messages is nil")
	}
	if len(stringMessages) != 0 {
		t.Errorf("string messages length = %d, want 0", len(stringMessages))
	}
	if cap(stringMessages) != 0 {
		t.Errorf("string messages capacity = %d, want 0", cap(stringMessages))
	}

	intMessages := CreateEmptyMessages[int]()
	if intMessages == nil {
		t.Error("int messages is nil")
	}
	if len(intMessages) != 0 {
		t.Errorf("int messages length = %d, want 0", len(intMessages))
	}

	interfaceMessages := CreateEmptyMessages[interface{}]()
	if interfaceMessages == nil {
		t.Error("interface messages is nil")
	}
	if len(interfaceMessages) != 0 {
		t.Errorf("interface messages length = %d, want 0", len(interfaceMessages))
	}
}

func TestPubSubDoc_FieldTypes(t *testing.T) {
	doc := PubSubDoc[string]{
		Messages:     []string{"msg1", "msg2"},
		CreationDate: 1234567890,
	}

	if len(doc.Messages) != 2 {
		t.Errorf("Messages length = %d, want 2", len(doc.Messages))
	}
	if doc.Messages[0] != "msg1" {
		t.Errorf("Messages[0] = %s, want msg1", doc.Messages[0])
	}
	if doc.Messages[1] != "msg2" {
		t.Errorf("Messages[1] = %s, want msg2", doc.Messages[1])
	}
	if doc.CreationDate != 1234567890 {
		t.Errorf("CreationDate = %d, want 1234567890", doc.CreationDate)
	}
}

func TestAssignmentDoc_Type(t *testing.T) {
	doc := make(AssignmentDoc)
	doc["group1"] = make(map[string]int64)
	doc["group1"]["instance1"] = 1234567890
	doc["group1"]["instance2"] = 1234567891

	doc["group2"] = make(map[string]int64)
	doc["group2"]["instance3"] = 1234567892

	if len(doc) != 2 {
		t.Errorf("AssignmentDoc length = %d, want 2", len(doc))
	}

	group1, exists := doc["group1"]
	if !exists {
		t.Error("group1 does not exist")
	}
	if len(group1) != 2 {
		t.Errorf("group1 length = %d, want 2", len(group1))
	}
	if group1["instance1"] != 1234567890 {
		t.Errorf("group1[instance1] = %d, want 1234567890", group1["instance1"])
	}

	group2, exists := doc["group2"]
	if !exists {
		t.Error("group2 does not exist")
	}
	if len(group2) != 1 {
		t.Errorf("group2 length = %d, want 1", len(group2))
	}
	if group2["instance3"] != 1234567892 {
		t.Errorf("group2[instance3] = %d, want 1234567892", group2["instance3"])
	}
}
