package upgrade

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const ftpDialTimeout = 20 * time.Second

var ftpKnownFiles = []string{
	"NPOS.exe",
	"NPOS.exe.sha256",
	"npos.exe",
	"npos.exe.sha256",
	"vanguard.exe",
	"vanguard.exe.sha256",
}

func ftpGet(addr, user, pass, dir, destDir string) error {
	return ftpGetMode(addr, user, pass, dir, destDir, true)
}

func ftpGetMode(addr, user, pass, dir, destDir string, passive bool) error {
	c, err := ftpConnect(addr)
	if err != nil {
		return fmt.Errorf("ftp connect %s: %w", addr, err)
	}
	defer c.close()
	if err := c.login(user, pass); err != nil {
		return fmt.Errorf("ftp login %s: %w", addr, err)
	}
	if _, err := c.cmd(200, "TYPE I"); err != nil {
		return fmt.Errorf("ftp type: %w", err)
	}
	if _, err := c.cmd(250, "CWD %s", dir); err != nil {
		if !isFTPCode(err, 200) {
			return fmt.Errorf("ftp cwd %s: %w", dir, err)
		}
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	names, err := c.listNames(passive)
	if err != nil {
		return fmt.Errorf("ftp list %s: %w", dir, err)
	}
	if len(names) == 0 {
		names = ftpKnownFiles
	}
	n := 0
	for _, name := range names {
		if strings.EqualFold(name, "tools.jar") || name == "." || name == ".." {
			continue
		}
		if err := c.retrTo(passive, name, filepath.Join(destDir, name)); err != nil {
			if isFTPCode(err, 550) {
				continue
			}
			return fmt.Errorf("ftp retr %s/%s: %w", dir, name, err)
		}
		n++
	}
	if n == 0 {
		return fmt.Errorf("no files in FTP folder %s", dir)
	}
	return nil
}

type ftpConn struct {
	c net.Conn
	r *bufio.Reader
}

func ftpConnect(addr string) (*ftpConn, error) {
	c, err := net.DialTimeout("tcp", addr, ftpDialTimeout)
	if err != nil {
		return nil, err
	}
	f := &ftpConn{c: c, r: bufio.NewReader(c)}
	if _, err := f.read(220); err != nil {
		c.Close()
		return nil, err
	}
	return f, nil
}

func (f *ftpConn) close() {
	_, _ = f.cmd(221, "QUIT")
	_ = f.c.Close()
}

func (f *ftpConn) login(user, pass string) error {
	if _, err := f.cmd(331, "USER %s", user); err != nil {
		if isFTPCode(err, 230) {
			return nil
		}
		return err
	}
	_, err := f.cmd(230, "PASS %s", pass)
	return err
}

func (f *ftpConn) cmd(want int, format string, args ...any) (string, error) {
	line := fmt.Sprintf(format, args...)
	_ = f.c.SetWriteDeadline(time.Now().Add(ftpDialTimeout))
	if _, err := io.WriteString(f.c, line+"\r\n"); err != nil {
		return "", err
	}
	return f.read(want)
}

func (f *ftpConn) read(want int) (string, error) {
	_ = f.c.SetReadDeadline(time.Now().Add(ftpDialTimeout))
	var msg strings.Builder
	code := 0
	for {
		line, err := f.r.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimRight(line, "\r\n")
		if len(line) < 3 {
			continue
		}
		n, err := strconv.Atoi(line[:3])
		if err != nil {
			continue
		}
		code = n
		rest := ""
		if len(line) > 4 {
			rest = line[4:]
		}
		if msg.Len() > 0 {
			msg.WriteByte('\n')
		}
		msg.WriteString(rest)
		if len(line) == 3 || line[3] == ' ' {
			break
		}
	}
	if want != 0 && code != want {
		return msg.String(), &ftpError{code: code, msg: msg.String()}
	}
	return msg.String(), nil
}

func (f *ftpConn) listNames(passive bool) ([]string, error) {
	data, err := f.dataConn(passive, "NLST")
	if err != nil {
		if isFTPCode(err, 500) || isFTPCode(err, 502) || isFTPCode(err, 550) {
			return nil, nil
		}
		return nil, err
	}
	_ = data.SetReadDeadline(time.Now().Add(2 * ftpDialTimeout))
	raw, err := io.ReadAll(data)
	_ = data.Close()
	if _, rerr := f.read(226); rerr != nil && err == nil {
		if isFTPCode(rerr, 250) {
			rerr = nil
		}
		err = rerr
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(strings.ReplaceAll(line, "\r", ""))
		if line == "" {
			continue
		}
		name := filepath.Base(strings.ReplaceAll(line, "\\", "/"))
		if name == "" || name == "." || name == ".." {
			continue
		}
		names = append(names, name)
	}
	return names, nil
}

func (f *ftpConn) retrTo(passive bool, name, dest string) error {
	data, err := f.dataConn(passive, "RETR %s", name)
	if err != nil {
		return err
	}
	out, err := os.Create(dest)
	if err != nil {
		data.Close()
		return err
	}
	_ = data.SetReadDeadline(time.Now().Add(5 * time.Minute))
	_, copyErr := io.Copy(out, data)
	out.Close()
	data.Close()
	_, doneErr := f.read(226)
	if copyErr != nil {
		_ = os.Remove(dest)
		return copyErr
	}
	if doneErr != nil && !isFTPCode(doneErr, 250) {
		_ = os.Remove(dest)
		return doneErr
	}
	return nil
}

func (f *ftpConn) dataConn(passive bool, format string, args ...any) (net.Conn, error) {
	if passive {
		return f.pasvCmd(format, args...)
	}
	return f.portCmd(format, args...)
}

func (f *ftpConn) pasvCmd(format string, args ...any) (net.Conn, error) {
	msg, err := f.cmd(227, "PASV")
	if err != nil {
		return nil, err
	}
	host, port, err := parsePASV(msg)
	if err != nil {
		return nil, err
	}
	data, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), ftpDialTimeout)
	if err != nil {
		return nil, err
	}
	if _, err := f.cmd(150, format, args...); err != nil {
		if !isFTPCode(err, 125) {
			data.Close()
			return nil, err
		}
	}
	return data, nil
}

func (f *ftpConn) portCmd(format string, args ...any) (net.Conn, error) {
	laddr, ok := f.c.LocalAddr().(*net.TCPAddr)
	if !ok || laddr.IP == nil {
		return nil, fmt.Errorf("ftp active: no local ip")
	}
	ip := laddr.IP.To4()
	if ip == nil {
		return nil, fmt.Errorf("ftp active: need ipv4")
	}
	ln, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: ip, Port: 0})
	if err != nil {
		return nil, err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	if _, err := f.cmd(200, "PORT %s", formatPORT(ip, port)); err != nil {
		ln.Close()
		return nil, err
	}
	type acc struct {
		c   net.Conn
		err error
	}
	ch := make(chan acc, 1)
	go func() {
		_ = ln.SetDeadline(time.Now().Add(ftpDialTimeout))
		c, err := ln.Accept()
		ln.Close()
		ch <- acc{c, err}
	}()
	if _, err := f.cmd(150, format, args...); err != nil && !isFTPCode(err, 125) {
		ln.Close()
		return nil, err
	}
	a := <-ch
	if a.err != nil {
		return nil, a.err
	}
	return a.c, nil
}

func parsePASV(msg string) (string, int, error) {
	start := strings.Index(msg, "(")
	end := strings.LastIndex(msg, ")")
	if start < 0 || end <= start {
		return "", 0, fmt.Errorf("invalid PASV response %q", msg)
	}
	parts := strings.Split(msg[start+1:end], ",")
	if len(parts) < 6 {
		return "", 0, fmt.Errorf("invalid PASV response %q", msg)
	}
	p1, err1 := strconv.Atoi(strings.TrimSpace(parts[4]))
	p2, err2 := strconv.Atoi(strings.TrimSpace(parts[5]))
	if err1 != nil || err2 != nil {
		return "", 0, fmt.Errorf("invalid PASV port %q", msg)
	}
	host := strings.Join([]string{
		strings.TrimSpace(parts[0]),
		strings.TrimSpace(parts[1]),
		strings.TrimSpace(parts[2]),
		strings.TrimSpace(parts[3]),
	}, ".")
	return host, p1*256 + p2, nil
}

func formatPORT(ip net.IP, port int) string {
	ip = ip.To4()
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d", ip[0], ip[1], ip[2], ip[3], port/256, port%256)
}

type ftpError struct {
	code int
	msg  string
}

func (e *ftpError) Error() string {
	if e.msg == "" {
		return fmt.Sprintf("ftp %d", e.code)
	}
	return fmt.Sprintf("ftp %d %s", e.code, e.msg)
}

func isFTPCode(err error, code int) bool {
	var fe *ftpError
	return errors.As(err, &fe) && fe.code == code
}

func isNetTimeout(err error) bool {
	if err == nil {
		return false
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "timeout") || strings.Contains(s, "timed out")
}
