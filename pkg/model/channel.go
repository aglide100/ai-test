package model

type RequestData struct {
    Data        string
    ResponseChan chan []byte
}