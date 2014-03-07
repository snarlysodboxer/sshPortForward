sshPortForward
===============

Golang SSH Port Forward, equivalent of `ssh -L localhost:LOCALPORT:localhost:REMOTEPORT user@server.com`

Usage:

`import "github.com/snarlysodboxer/sshPortForward"`

    sshPortForward.ConnectAndForward(userNameString serverAddrString localAddrString remoteAddrString privateKeyPathString)
