package main

type Tab struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Host     string `json:"host"`
	URL      string `json:"url"`
	WindowID int    `json:"windowId"`
}
