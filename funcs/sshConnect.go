package funcs

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/mySSH/g"
	"golang.org/x/crypto/ssh"
)

func connect(user, password, host string, port int) (*ssh.Session, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		config       ssh.Config
		session      *ssh.Session
		err          error
	)
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))
	config = ssh.Config{
		Ciphers:      []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"},
		KeyExchanges: []string{"diffie-hellman-group-exchange-sha1", "diffie-hellman-group1-sha1", "diffie-hellman-group-exchange-sha256"},
	}

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		Config:  config,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	addr = fmt.Sprintf("%s:%d", host, port)
	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	if session, err = client.NewSession(); err != nil {
		return nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return nil, err
	}
	return session, nil
}

func Dossh(username, password, host string, port int, cmds string, ch chan g.SSHResult) {
	chSSH := make(chan g.SSHResult)
	timeout := 30
	go dosshSession(username, password, host, port, cmds, chSSH)
	var res g.SSHResult
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		res.Host = host
		res.Success = false
		res.Result = ("SSH run timeout: " + strconv.Itoa(timeout) + " second.")
		ch <- res
	case res = <-chSSH:
		ch <- res
	}
	return
}

func dosshSession(username, password, host string, port int, cmds string, ch chan g.SSHResult) {
	session, err := connect(username, password, host, port)
	var sshResult g.SSHResult
	sshResult.Host = host
	if err != nil {
		sshResult.Success = false
		sshResult.Result = fmt.Sprintf("<%s>", err.Error())
		ch <- sshResult
		return
	}
	defer session.Close()

	stdinBuf, _ := session.StdinPipe()
	var outbt, errbt bytes.Buffer

	session.Stdout = &outbt
	session.Stderr = &errbt

	err = session.Shell()

	if err != nil {
		sshResult.Success = false
		sshResult.Result = fmt.Sprintf("<%s>", err.Error())
		ch <- sshResult
		return
	}
	stdinBuf.Write([]byte(cmds))
	session.Wait()
	if errbt.String() != "" {
		sshResult.Success = false
		sshResult.Result = errbt.String()
		ch <- sshResult
	} else {
		sshResult.Success = true
		sshResult.Result = outbt.String()
		ch <- sshResult
	}
	return
}

func dosshRun(username, password, host string, port int, cmds string, ch chan g.SSHResult) {
	session, err := connect(username, password, host, port)
	var sshResult g.SSHResult
	sshResult.Host = host
	if err != nil {
		sshResult.Success = false
		sshResult.Result = fmt.Sprintf("<%s>", err.Error())
		ch <- sshResult
		return
	}
	defer session.Close()

	var outbt, errbt bytes.Buffer
	session.Stdout = &outbt

	session.Stderr = &errbt
	err = session.Run(cmds)
	if err != nil {
		sshResult.Success = false
		sshResult.Result = fmt.Sprintf("<%s>", err.Error())
		ch <- sshResult
		return
	}

	if errbt.String() != "" {
		sshResult.Success = false
		sshResult.Result = errbt.String()
		ch <- sshResult
	} else {
		sshResult.Success = true
		sshResult.Result = outbt.String()
		ch <- sshResult
	}

	return
}
