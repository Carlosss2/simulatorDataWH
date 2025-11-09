package models

import "time"


type StatusUpdate struct {
	Timestamp     time.Time
	IsConnected   bool    
	IsRunning     bool    
	MessagesSent  int64   
	Error         string  
}

const (
	CmdStart = "START" 
	CmdStop  = "STOP"  
)