package protocol

import errors

const (
   /*memcached storage command*/
   MC_SET = iota
   MC_ADD
   MC_REPL
   MC_APPE
   MC_PREP
   MC_CAS
   /*memcached retrieval command*/
   MC_GET
   MC_GETS
   /*memcached deletion command*/
   MC_DEL
   /*memcached inc/dec command*/
   MC_INCR
   MC_DECR
   /*memcached touch command*/
   MC_TOUCH
)
/* means the client sent a nonexistent command name. */
var MC_ERR_CMD_NOEXIST = "ERROR\r\n"
/* means some sort of client error in the input line. */
var MC_ERR_CLIENT_ERR = "CLIENT_ERROR"
/* means some sort of server error prevents the server 
from carrying out the command. */
var MC_ERR_SERVER_ERR = "SERVER_ERROR"

func ParseMcRequest( 
