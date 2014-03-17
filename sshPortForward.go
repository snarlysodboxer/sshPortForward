package sshPortForward

import (
	"code.google.com/p/go.crypto/ssh"
	"io"
	"log"
	"net"
  "io/ioutil"
)

type Addresses struct {
  SSHUserString         string
  ServerAddrString      string
  RemoteAddrString      string
  LocalAddrString       string
  PrivateKeyPathString  string
}

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

func forward(localConn net.Conn, config *ssh.ClientConfig, serverAddrString, remoteAddrString string) {
	// Setup sshClientConn (type *ssh.ClientConn)
	sshClientConn, err := ssh.Dial("tcp", serverAddrString, config)
	if err != nil {
		log.Printf("ssh.Dial failed: %s", err)
    return
	}
  //defer sshClientConn.Close()

	// Setup sshConn (type net.Conn)
	sshConn, err := sshClientConn.Dial("tcp", remoteAddrString)
	if err != nil {
		log.Printf("sshClientConn.Dial failed: %s", err)
    return
	}
  //defer sshConn.Close()

	// Copy localConn.Reader to sshConn.Writer
	go func() {
		_, err = io.Copy(sshConn, localConn)
		if err != nil {
			log.Printf("io.Copy from local to remote failed: %v", err)
		}
	}()

	// Copy sshConn.Reader to localConn.Writer
	go func() {
		_, err = io.Copy(localConn, sshConn)
		if err != nil {
			log.Printf("io.Copy from remote to local failed: %v", err)
		}
	}()
}

func ConnectAndForward(addresses Addresses) {
  // Load id_rsa file
  keychain := new(keyChain)
  err := keychain.loadPEM(addresses.PrivateKeyPathString)
  if err != nil {
    log.Fatalf("Cannot load key: %v", err)
  }

	// Setup SSH config (type *ssh.ClientConfig)
	config := &ssh.ClientConfig{
		User: addresses.SSHUserString,
		Auth: []ssh.ClientAuth{
			ssh.ClientAuthKeyring(keychain),
		},
	}

	// Setup localListener (type net.Listener)
	localListener, err := net.Listen("tcp", addresses.LocalAddrString)
	if err != nil {
		log.Printf("net.Listen failed: %v", err)
    // Don't setup a connection where one already exists:
    return
	}
  defer localListener.Close()

	for {
		// Setup localConn (type net.Conn)
		localConn, err := localListener.Accept()
		if err != nil {
			log.Printf("listen.Accept failed: %v", err)
		}
    defer localConn.Close()
		go forward(localConn, config, addresses.ServerAddrString, addresses.RemoteAddrString)
	}
}
