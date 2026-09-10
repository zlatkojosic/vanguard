package print

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

type Settings struct {
	UseFiscal         string
	UseSerial         string
	FiscalType        string
	SerialType        string
	PrinterName       string
	PrinterFolder     string
	FileExt           string
	ComPort           string
	RowLen            int
	LeftMargin        int
	TopMargin         int
	PageWidth         int
	PageScale         float64
}

type Job struct {
	Settings Settings
	Template *Template
	Params   Params
	Content  string
}

func (j *Job) Rendered() string {
	if j.Content != "" {
		return j.Content
	}
	row := j.Settings.RowLen
	if row <= 0 {
		row = 30
	}
	j.Content = Render(j.Template, j.Params, row)
	return j.Content
}

func (j *Job) Execute() error {
	if j.Settings.UseFiscal == "true" {
		return writeFiscal(j)
	}
	if j.Settings.UseSerial == "true" {
		return writeSerial(j)
	}
	return writeOS(j)
}

func writeOS(j *Job) error {
	content := j.Rendered()
	tmp, err := os.CreateTemp("", "vanguard-print-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()
	switch runtime.GOOS {
	case "darwin":
		args := []string{tmp.Name()}
		if j.Settings.PrinterName != "" {
			args = []string{"-d", j.Settings.PrinterName, tmp.Name()}
		}
		return exec.Command("lpr", args...).Run()
	case "linux":
		args := []string{tmp.Name()}
		if j.Settings.PrinterName != "" {
			args = []string{"-P", j.Settings.PrinterName, tmp.Name()}
		}
		return exec.Command("lp", args...).Run()
	case "windows":
		return exec.Command("notepad", "/p", tmp.Name()).Run()
	default:
		return fmt.Errorf("unsupported OS print: %s", runtime.GOOS)
	}
}

func writeSerial(j *Job) error {
	portName := j.Settings.ComPort
	if portName == "" {
		return fmt.Errorf("COM port nije izabran")
	}
	mode := &serial.Mode{BaudRate: 9600, DataBits: 8, Parity: serial.NoParity, StopBits: serial.OneStopBit}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return err
	}
	defer port.Close()

	content := j.Rendered()
	enc := charmap.CodePage852
	if j.Settings.SerialType == "Secunda" {
		encoded, err := enc.NewEncoder().Bytes([]byte(content))
		if err != nil {
			encoded = []byte(content)
		}
		lines := strings.Split(string(encoded), "\n")
		for _, line := range lines {
			payload := append([]byte{26, 0, 255}, []byte(line)...)
			if _, err := port.Write(payload); err != nil {
				return err
			}
		}
		_, err = port.Write([]byte{26, 1, 255})
		return err
	}
	_, err = port.Write([]byte(content))
	return err
}

var fiscalCounter struct {
	sync.Mutex
	n int
}

func nextFiscalID() int {
	fiscalCounter.Lock()
	defer fiscalCounter.Unlock()
	fiscalCounter.n++
	if fiscalCounter.n > 99 {
		fiscalCounter.n = 1
	}
	return fiscalCounter.n
}

func writeFiscal(j *Job) error {
	folder := j.Settings.PrinterFolder
	if folder == "" {
		return fmt.Errorf("folder štampača je prazan")
	}
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return err
	}
	content := j.Rendered()
	typ := strings.ToLower(j.Settings.FiscalType)
	body, name, encName := fiscalPayload(typ, content, j.Settings, nextFiscalID())
	enc := encodingFor(encName)
	encoded, err := enc.NewEncoder().Bytes([]byte(body))
	if err != nil {
		encoded = []byte(body)
	}
	return os.WriteFile(filepath.Join(folder, name), encoded, 0o644)
}

func fiscalPayload(typ, content string, s Settings, id int) (body, filename, encoding string) {
	ext := s.FileExt
	if ext == "" {
		ext = "inp"
	}
	switch typ {
	case "elink":
		var b strings.Builder
		b.WriteString("L,1,______,_,__;\n")
		for _, line := range strings.Split(content, "\n") {
			fmt.Fprintf(&b, "I,1,______,_,__;%s;\n", line)
		}
		b.WriteString("H,1,______,_,__;\n")
		return b.String(), fmt.Sprintf("snd.%d%d.%s", time.Now().UnixMilli(), id, ext), "cp1250"
	case "tring":
		xml := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
   <Zahtjev xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
     <BrojZahtjeva>%d</BrojZahtjeva>
     <VrstaZahtjeva>6</VrstaZahtjeva>
     <Parametri>
        <Parametar>
           <Naziv>Text</Naziv>
           <Vrijednost>%s</Vrijednost>
        </Parametar>
     </Parametri>
  </Zahtjev>
`, id, xmlEscape(content))
		return xml, fmt.Sprintf("snd.%d.xml", id), "UTF-8"
	case "tringt202":
		var b strings.Builder
		b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
         <USER_TEXT>
`)
		for _, line := range strings.Split(content, "\n") {
			fmt.Fprintf(&b, `            <DATA TXT="%s"/>`+"\n", xmlEscape(line))
		}
		b.WriteString("         </USER_TEXT>\n")
		return b.String(), fmt.Sprintf("snd.%d.xml", id), "UTF-8"
	case "tremol":
		xml := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8" ?>
<TremolFpServer Command="NonFiscal">
  <Print>%s</Print>
</TremolFpServer>        
`, xmlEscape(content))
		if ext == "inp" {
			ext = "xml"
		}
		return xml, fmt.Sprintf("snd.%d.%s", id, ext), "UTF-8"
	case "fp&tm", "fptm":
		var b strings.Builder
		b.WriteString("38,1,______,_,__;\r\n")
		for _, line := range strings.Split(content, "\n") {
			fmt.Fprintf(&b, "42,1,______,_,__;%s;\r\n", line)
		}
		b.WriteString("39,1,______,_,__;\r\n")
		return b.String(), fmt.Sprintf("snd.%d%d.%s", time.Now().UnixMilli(), id, ext), "cp1250"
	case "mep1000":
		var b strings.Builder
		b.WriteByte('{')
		for i, line := range strings.Split(content, "\n") {
			if i > 0 {
				b.WriteByte('\n')
			}
			fmt.Fprintf(&b, "P,1,______,_,__;%s;", line)
		}
		b.WriteString("}\n")
		return b.String(), fmt.Sprintf("snd.%d%d.%s", time.Now().UnixMilli(), id, ext), "cp1250"
	default:
		return content, fmt.Sprintf("snd.%d.txt", id), "UTF-8"
	}
}

func xmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}

func encodingFor(name string) encoding.Encoding {
	switch strings.ToLower(name) {
	case "cp1250":
		return charmap.Windows1250
	case "cp852":
		return charmap.CodePage852
	default:
		return unicode.UTF8
	}
}

func ListPrinters() []string {
	var out []byte
	var err error
	switch runtime.GOOS {
	case "darwin", "linux":
		out, err = exec.Command("lpstat", "-a").Output()
	case "windows":
		out, err = exec.Command("wmic", "printer", "get", "name").Output()
	}
	if err != nil {
		return nil
	}
	var names []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(strings.ToLower(line), "name") {
			continue
		}
		if fields := strings.Fields(line); len(fields) > 0 {
			names = append(names, fields[0])
		}
	}
	return names
}
