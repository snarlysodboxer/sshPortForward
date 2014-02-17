package main

// Forward from local port 9000 to remote port 9999

import (
	"code.google.com/p/go.crypto/ssh"
	"io"
	"log"
	"net"
  "io/ioutil"
)

var (
	username         = "root"
	serverAddrString = "192.168.1.100:22"
	localAddrString  = "localhost:9000"
	remoteAddrString = "localhost:9999"
)

type keyChain struct {
  keys []ssh.Signer
}

func (keychain *keyChain) Key(number int) (ssh.PublicKey, error) {
  if number < 0 || number >= len(keychain.keys) {
    return nil, nil
  }
  return keychain.keys[number].PublicKey(), nil
}

func (keychain *keyChain) Sign(number int, rand io.Reader, data []byte) (sig []byte, err error) {
  return keychain.keys[number].Sign(rand, data)
}

func (keychain *keyChain) add(key ssh.Signer) {
  keychain.keys = append(keychain.keys, key)
}

func (keychain *keyChain) loadPEM(file string) error {
  buf, err := ioutil.ReadFile(file)
  if err != nil {
    return err
  }
  key, err := ssh.ParsePrivateKey(buf)
  if err != nil {
    return err
  }
  keychain.add(key)
  return nil
}

func forward(localConn net.Conn, config *ssh.ClientConfig) {
	// Setup sshClientConn (type *ssh.ClientConn)
	sshClientConn, err := ssh.Dial("tcp", serverAddrString, config)
	if err != nil {
		log.Fatalf("ssh.Dial failed: %s", err)
	}

	// Setup sshConn (type net.Conn)
	sshConn, err := sshClientConn.Dial("tcp", remoteAddrString)
	if err != nil {
		log.Fatalf("sshClientConn.Dial failed: %s", err)
	}

	// Copy localConn.Reader to sshConn.Writer
	go func() {
		_, err = io.Copy(sshConn, localConn)
		if err != nil {
			log.Fatalf("io.Copy failed: %v", err)
		}
	}()

	// Copy sshConn.Reader to localConn.Writer
	go func() {
		_, err = io.Copy(localConn, sshConn)
		if err != nil {
			log.Fatalf("io.Copy failed: %v", err)
		}
	}()
}

func main() {
  // Load id_rsa file
  keychain := new(keyChain)
  err := keychain.loadPEM("/home/myuser/.ssh/id_rsa")
  if err != nil {
    log.Fatalf("Cannot load key: %v", err)
  }

	// Setup SSH config (type *ssh.ClientConfig)
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.ClientAuth{
			ssh.ClientAuthKeyring(keychain),
		},
	}

	// Setup localListener (type net.Listener)
	localListener, err := net.Listen("tcp", localAddrString)
	if err != nil {
		log.Fatalf("net.Listen failed: %v", err)
	}

	for {
		// Setup localConn (type net.Conn)
		localConn, err := localListener.Accept()
		if err != nil {
			log.Fatalf("listen.Accept failed: %v", err)
		}
		go forward(localConn, config)
	}
}
