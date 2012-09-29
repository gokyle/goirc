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

/*
   This package is provided under a dual ISC / public domain license.
   The public domain license is the one applicable to the user of this
   code; you are free to choose whichever affords the maximum freedom
   to you.
  
   --------------------------------------------------------------------
   The ISC license:
  
   Copyright (c) 2012 Kyle Isom <kyle@tyrfingr.is>
  
   Permission to use, copy, modify, and distribute this software for any
   purpose with or without fee is hereby granted, provided that the 
   above copyright notice and this permission notice appear in all 
   copies.
  
   THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL 
   WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED 
   WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE 
   AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL
   DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA
   OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER 
   TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR 
   PERFORMANCE OF THIS SOFTWARE.
   --------------------------------------------------------------------
 */
package goirc
