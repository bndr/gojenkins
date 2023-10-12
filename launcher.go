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

type sshLauncher struct {
	XMLName              xml.Name      `xml:"launcher" json:"-"`
	Class                LauncherClass `xml:"-" json:"-"`
	Host                 string        `xml:"host" json:"host"`
	Port                 int           `xml:"port" json:"port"`
	CredentialsId        string        `xml:"credentialsId" json:"credentialsId"`
	LaunchTimeoutSeconds int           `xml:"launchTimeoutSeconds" json:"launchTimeoutSeconds"`
	MaxNumRetries        int           `xml:"maxNumRetries" json:"maxNumRetries"`
	RetryWaitTime        int           `xml:"retryWaitTime" json:"retryWaitTime"`
	JvmOptions           string        `xml:"jvmOptions" json:"jvmOptions"`
	JavaPath             string        `xml:"javaPath" json:"JavaPath"`
	PrefixStartSlaveCmd  string        `xml:"prefixStartSlaveCmd" json:"prefixStartSlaveCmd"`
	SuffixStartSlaveCmd  string        `xml:"suffixStartSlaveCmd" json:"suffixStartSlaveCmd"`
}

type WorkDirSettings struct {
	Disabled               bool   `xml:"disabled" json:"disabled"`
	InternalDir            string `xml:"internalDir" json:"internalDir"`
	FailIfWorkDirIsMissing bool   `xml:"failIfWorkDirIsMissing" json:"failIfWorkDirIsMissing"`
}
type jnlpLauncher struct {
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
		var sshLauncher sshLauncher
		if err := d.DecodeElement(&sshLauncher, &start); err != nil {
			return err
		}
		sshLauncher.Class = sshLauncher.GetClass()
		c.Launcher = &sshLauncher
	case JNLPLauncherClass:
		var jnlpLauncher jnlpLauncher
		if err := d.DecodeElement(&jnlpLauncher, &start); err != nil {
			return err
		}
		jnlpLauncher.Class = jnlpLauncher.GetClass()
		c.Launcher = &jnlpLauncher
	default:
		return errors.New("unknown launcher class")
	}
	return nil
}

func (c *CustomLauncher) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	switch l := c.Launcher.(type) {
	case *sshLauncher:
		start.Name.Local = "launcher"
		start.Attr = []xml.Attr{{Name: xml.Name{Local: "class"}, Value: string(SSHLauncherClass)}}
		return e.EncodeElement(l, start)
	case *jnlpLauncher:
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

func (s *sshLauncher) GetClass() LauncherClass {
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
	PrefixStartSlaveCmd string,
	SuffixStartSlaveCmd string) *sshLauncher {
	return &sshLauncher{Class: SSHLauncherClass}
}

// Returns the defaults that Jenkins fills out when no options are given.
func DefaultSSHLauncher() *sshLauncher {
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

func (j *jnlpLauncher) GetClass() LauncherClass {
	return JNLPLauncherClass
}

func NewJNLPLauncher(webSocket bool, w *WorkDirSettings) *jnlpLauncher {
	return &jnlpLauncher{Class: JNLPLauncherClass,
		WorkDirSettings: w,
		WebSocket:       webSocket}
}

func DefaultJNLPLauncher() *jnlpLauncher {
	return NewJNLPLauncher(false, &WorkDirSettings{
		Disabled:               false,
		InternalDir:            "remoting",
		FailIfWorkDirIsMissing: false,
	})
}
