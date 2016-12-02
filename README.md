# Network Security Projects
##Implementation of CSC474, Network Security projects in Go
This is a repo that contanins my implementation of some of our network security projects in Go. Many of them implement crypto libraries to allow secure network communication for the application being used.
Our projects were required to either be implemented in Python or C, so instead of just posting those up here (which felt wrong because it makes copying easy) I decided to just reimplement them in my favorite language. 
So far this repo only contains a few of the projects as I am working diligently to port the rest of them over to Go. 

##Current Programs in this repo:
###UnecryptedIM.go

####A basic messaging client over TCP that can act as a server or client

```
usage: ./UnencryptedIM [-s | -c C]

A P2P IM service.

Mutually Exclusive Arguments:
  -s              Start an IM server
  -c C            Connect to the specified IM server
```

###EncryptedIM.go

####An encrypted messaging client that uses two user preshared authenticty and confidentiality keys

```
usage: ./EncryptedIM [-c HOSTNAME] [-s] [-confkey CONFIDENTIALITY KEY]
                      [-authkey AUTHENTICITY KEY] [-p PORT]

A P2P IM service.

optional arguments:
  -c HOSTNAME           Host to connect to
  -s                    Run as server (on port 9999)
  -confkey CONFIDENTIALITY KEY
                        Key used in encryption
  -authkey AUTHENTICITY KEY
                        Key with which HMAC is computed
```
