/*
   Package goirc implements a basic IRC client. goirc is implemented
   around the Irc structure.

   An IRC configuration can be read from a JSON file; see the NewIrc 
   function. For example,
   
   irc, err := goirc.NewIrc("config.json")

   goirc implements a number of commands that act on the Irc structure;
   an IRC session is initiated with the Connect() method. PRIVMSGs can
   be sent with the Msg method, and data may be read from the session
   using the Recv method. 

   The Pong method is designed to be used as a goroutine; when a line
   is returned from Recv(), the pingRegex can be used to determine
   whether incoming data is a PING request:

   func isPing(msg string) (bool, string) {
	    if !pingRegex.MatchString(msg) {
	    	    return false, ""
	    }
	    return pingRegex.ReplaceAllString(msg, "$1"), true
    }

    if pinged, sender := isPing(msg); pinged {
            go client.Pong(sender)
    }
*/
package goirc
