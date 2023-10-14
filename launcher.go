package gojenkins

import (
	"encoding/xml"
	"errors"
)

type LauncherClass string

const (
	JNLPLauncherClass LauncherClass = "hudson.slaves.JNLPLauncher"
	SSHLauncherClass  LauncherClass = "hudson.plugins.sshslaves.SSHLauncher"
)

type Launcher interface {
	// Give a fake interface to implement so
	// You can't pass anything into the Launcher interface.
	GetClass() LauncherClass
}

type CustomLauncher struct {
	XMLName  xml.Name      `xml:"launcher" json:"-"`
	Launcher Launcher      `xml:"-"`
	Class    LauncherClass `xml:"class,attr" json:"$class"`
}

type SSHLauncher struct {
	XMLName              xml.Name      `xml:"launcher" json:"-"`
	Class                LauncherClass `xml:"-" json:"-"`
	Host                 string        `xml:"host" json:"host"`
	Port                 int           `xml:"port" json:"port"`
	CredentialsId        string        `xml:"credentialsId" json:"credentialsId"`
	LaunchTimeoutSeconds int           `xml:"launchTimeoutSeconds" json:"launchTimeoutSeconds"`
	MaxNumRetries        int           `xml:"maxNumRetries" json:"maxNumRetries"`
	RetryWaitTime        int           `xml:"retryWaitTime" json:"retryWaitTime"`
	JvmOptions           string        `xml:"jvmOptions" json:"jvmOptions"`
	JavaPath             string        `xml:"javaPath" json:"javaPath"`
	PrefixStartSlaveCmd  string        `xml:"prefixStartSlaveCmd" json:"prefixStartSlaveCmd"`
	SuffixStartSlaveCmd  string        `xml:"suffixStartSlaveCmd" json:"suffixStartSlaveCmd"`
}

type WorkDirSettings struct {
	Disabled               bool   `xml:"disabled" json:"disabled"`
	InternalDir            string `xml:"internalDir" json:"internalDir"`
	FailIfWorkDirIsMissing bool   `xml:"failIfWorkDirIsMissing" json:"failIfWorkDirIsMissing"`
}
type JNLPLauncher struct {
	XMLName         xml.Name         `xml:"launcher" json:"-"`
	Class           LauncherClass    `xml:"-" json:"-"`
	WorkDirSettings *WorkDirSettings `xml:"workDirSettings" json:"workDirSettings"`
	WebSocket       bool             `xml:"webSocket" json:"webSocket"`
}

func (c *CustomLauncher) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	c.Class = ""
	for _, attr := range start.Attr {
		if attr.Name.Local == "class" {
			c.Class = LauncherClass(attr.Value)
			break
		}
	}
	switch c.Class {
	case SSHLauncherClass:
		var SSHLauncher SSHLauncher
		if err := d.DecodeElement(&SSHLauncher, &start); err != nil {
			return err
		}
		SSHLauncher.Class = SSHLauncher.GetClass()
		c.Launcher = &SSHLauncher
	case JNLPLauncherClass:
		var JNLPLauncher JNLPLauncher
		if err := d.DecodeElement(&JNLPLauncher, &start); err != nil {
			return err
		}
		JNLPLauncher.Class = JNLPLauncher.GetClass()
		c.Launcher = &JNLPLauncher
	default:
		return errors.New("unknown launcher class")
	}
	return nil
}

func (c *CustomLauncher) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	switch l := c.Launcher.(type) {
	case *SSHLauncher:
		start.Name.Local = "launcher"
		start.Attr = []xml.Attr{{Name: xml.Name{Local: "class"}, Value: string(SSHLauncherClass)}}
		return e.EncodeElement(l, start)
	case *JNLPLauncher:
		start.Name.Local = "launcher"
		start.Attr = []xml.Attr{{Name: xml.Name{Local: "class"}, Value: string(JNLPLauncherClass)}}
		return e.EncodeElement(l, start)
	case nil:
		// Encodes empty element. With just the class.
		start.Name.Local = "launcher"
		start.Attr = []xml.Attr{{Name: xml.Name{Local: "class"}, Value: string(c.Class)}}
		if err := e.EncodeToken(start); err != nil {
			return err
		}
		return e.EncodeToken(xml.EndElement{Name: start.Name})
	default:
		return errors.New("unsupported launcher type")
	}
}

func (s *SSHLauncher) GetClass() LauncherClass {
	return SSHLauncherClass
}

func NewSSHLauncher(
	host string,
	port int,
	credentialsId string,
	launchTimeout int,
	maxRetries int,
	retryWaitTime int,
	jvmOptions string,
	javaPath string,
	prefixStartSlaveCmd string,
	suffixStartSlaveCmd string) *SSHLauncher {
	return &SSHLauncher{
		Class:                SSHLauncherClass,
		Host:                 host,
		Port:                 port,
		CredentialsId:        credentialsId,
		LaunchTimeoutSeconds: launchTimeout,
		MaxNumRetries:        maxRetries,
		RetryWaitTime:        retryWaitTime,
		JvmOptions:           jvmOptions,
		JavaPath:             javaPath,
		PrefixStartSlaveCmd:  prefixStartSlaveCmd,
		SuffixStartSlaveCmd:  suffixStartSlaveCmd}
}

// Returns the defaults that Jenkins fills out when no options are given.
func DefaultSSHLauncher() *SSHLauncher {
	return NewSSHLauncher(
		"",
		22,
		"",
		60,
		0,
		0,
		"",
		"",
		"",
		"",
	)
}

func (j *JNLPLauncher) GetClass() LauncherClass {
	return JNLPLauncherClass
}

func NewJNLPLauncher(webSocket bool, w *WorkDirSettings) *JNLPLauncher {
	return &JNLPLauncher{Class: JNLPLauncherClass,
		WorkDirSettings: w,
		WebSocket:       webSocket}
}

func DefaultJNLPLauncher() *JNLPLauncher {
	return NewJNLPLauncher(false, &WorkDirSettings{
		Disabled:               false,
		InternalDir:            "remoting",
		FailIfWorkDirIsMissing: false,
	})
}
