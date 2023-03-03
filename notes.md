Read Request

      client                                           server
      -------------------------------------------------------
      |1|foofile|0|octet|0|blksize|0|1432|0|  -->               RRQ  // Sequential (receives request, blocks, sends back to client)
                                    <--  |6|blksize|0|1432|0|   OACK // Seuqential (client receives acknowledge that request is accepted)
                             <--  |3|0| 1432 octets of data |   DATA
                             <--  |3|1| 1432 octets of data |   DATA
                             <--  |3|2| 1432 octets of data |   DATA
                             <--  |3|3| 1432 octets of data |   DATA