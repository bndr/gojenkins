package gojenkins

import (
	"encoding/xml"
	"io"
	"strings"
)

// NodeProperty represents a Jenkins node property interface
type NodeProperty interface {
	// GetClass returns the Java class name for the node property
	GetClass() string
}

// NodeProperties is a wrapper for multiple node properties
type NodeProperties struct {
	XMLName    xml.Name       `xml:"nodeProperties"`
	Properties []NodeProperty `xml:",any"`
}

// CustomNodeProperty wraps a node property with its class attribute
type CustomNodeProperty struct {
	XMLName  xml.Name     `xml:",any"`
	Property NodeProperty `xml:"-"`
	Class    string       `xml:"class,attr"`
}

// EnvironmentVariablesNodeProperty allows setting environment variables on a node
type EnvironmentVariablesNodeProperty struct {
	XMLName xml.Name             `xml:"hudson.slaves.EnvironmentVariablesNodeProperty"`
	Class   string               `xml:"-"`
	EnvVars EnvironmentVariables `xml:"envVars"`
}

type EnvironmentVariables struct {
	XMLName xml.Name `xml:"envVars"`
	Tree    []EnvVar `xml:"-"`
}

type EnvVar struct {
	Key   string
	Value string
}

// encodeTreeMapHeader encodes the TreeMap header with comparator for Jenkins custom serialization
func encodeTreeMapHeader(enc *xml.Encoder) error {
	// <default><comparator class="java.lang.String$CaseInsensitiveComparator"/></default>
	type Comparator struct {
		XMLName xml.Name `xml:"comparator"`
		Class   string   `xml:"class,attr"`
	}
	type Default struct {
		XMLName    xml.Name   `xml:"default"`
		Comparator Comparator `xml:"comparator"`
	}
	return enc.Encode(Default{
		Comparator: Comparator{Class: "java.lang.String$CaseInsensitiveComparator"},
	})
}

// encodeEnvVarEntries encodes key-value pairs as <string> elements
func encodeEnvVarEntries(enc *xml.Encoder, entries []EnvVar) error {
	type String struct {
		XMLName xml.Name `xml:"string"`
		Value   string   `xml:",chardata"`
	}
	for _, entry := range entries {
		if err := enc.Encode(String{Value: entry.Key}); err != nil {
			return err
		}
		if err := enc.Encode(String{Value: entry.Value}); err != nil {
			return err
		}
	}
	return nil
}

// MarshalXML custom marshaler for EnvironmentVariables to match Jenkins' custom serialization format
func (e EnvironmentVariables) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	type UnserializableParents struct {
		XMLName xml.Name `xml:"unserializable-parents"`
	}
	type Int struct {
		XMLName xml.Name `xml:"int"`
		Value   int      `xml:",chardata"`
	}
	
	start.Name.Local = "envVars"
	start.Attr = []xml.Attr{{Name: xml.Name{Local: "serialization"}, Value: "custom"}}
	if err := enc.EncodeToken(start); err != nil {
		return err
	}
	
	if err := enc.Encode(UnserializableParents{}); err != nil {
		return err
	}
	
	// <tree-map> needs manual handling due to mixed content
	treeMapStart := xml.StartElement{Name: xml.Name{Local: "tree-map"}}
	if err := enc.EncodeToken(treeMapStart); err != nil {
		return err
	}
	
	if err := encodeTreeMapHeader(enc); err != nil {
		return err
	}
	
	if err := enc.Encode(Int{Value: len(e.Tree)}); err != nil {
		return err
	}
	
	if err := encodeEnvVarEntries(enc, e.Tree); err != nil {
		return err
	}
	
	if err := enc.EncodeToken(xml.EndElement{Name: treeMapStart.Name}); err != nil {
		return err
	}
	return enc.EncodeToken(xml.EndElement{Name: start.Name})
}

// UnmarshalXML custom unmarshaler for EnvironmentVariables to handle Jenkins' custom serialization format
func (e *EnvironmentVariables) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var strings []string
	var count int
	countRead := false
	
	for {
		token, err := d.Token()
		if err != nil {
			return err
		}
		
		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "int" {
				if err := d.DecodeElement(&count, &t); err != nil {
					return err
				}
				countRead = true
			} else if t.Name.Local == "string" && countRead {
				var s string
				if err := d.DecodeElement(&s, &t); err != nil {
					return err
				}
				strings = append(strings, s)
			} else if t.Name.Local == "default" || t.Name.Local == "comparator" {
				// Skip these metadata elements
				if err := d.Skip(); err != nil {
					return err
				}
			}
			// Don't skip tree-map, unserializable-parents - just continue parsing their children
		case xml.EndElement:
			if t.Name.Local == "envVars" {
				// Parse string pairs into EnvVar entries
				for i := 0; i+1 < len(strings); i += 2 {
					e.Tree = append(e.Tree, EnvVar{
						Key:   strings[i],
						Value: strings[i+1],
					})
				}
				return nil
			}
		}
	}
}

func (e *EnvironmentVariablesNodeProperty) GetClass() string {
	return "hudson.slaves.EnvironmentVariablesNodeProperty"
}

// NewEnvironmentVariablesNodeProperty creates a new environment variables node property
func NewEnvironmentVariablesNodeProperty(envVars map[string]string) *EnvironmentVariablesNodeProperty {
	vars := make([]EnvVar, 0, len(envVars))
	for k, v := range envVars {
		vars = append(vars, EnvVar{Key: k, Value: v})
	}
	return &EnvironmentVariablesNodeProperty{
		Class: "hudson.slaves.EnvironmentVariablesNodeProperty",
		EnvVars: EnvironmentVariables{
			Tree: vars,
		},
	}
}

// ToolLocationNodeProperty allows specifying tool locations on a node
type ToolLocationNodeProperty struct {
	XMLName   xml.Name       `xml:"hudson.tools.ToolLocationNodeProperty"`
	Class     string         `xml:"-"`
	Locations []ToolLocation `xml:"locations>hudson.tools.ToolLocationNodeProperty_-ToolLocation"`
}

type ToolLocation struct {
	Type string `xml:"type"`
	Name string `xml:"name"`
	Home string `xml:"home"`
}

func (t *ToolLocationNodeProperty) GetClass() string {
	return "hudson.tools.ToolLocationNodeProperty"
}

// NewToolLocationNodeProperty creates a new tool location node property
// The map key format should be "toolType:toolName", e.g., "hudson.plugins.git.GitTool$DescriptorImpl:Default"
func NewToolLocationNodeProperty(locations map[string]string) *ToolLocationNodeProperty {
	locs := make([]ToolLocation, 0, len(locations))
	for k, v := range locations {
		// Split the key into type and name (format: "type:name")
		parts := strings.SplitN(k, ":", 2)
		toolType := parts[0]
		toolName := "Default"
		if len(parts) > 1 {
			toolName = parts[1]
		}
		locs = append(locs, ToolLocation{
			Type: toolType,
			Name: toolName,
			Home: v,
		})
	}
	return &ToolLocationNodeProperty{
		Class:     "hudson.tools.ToolLocationNodeProperty",
		Locations: locs,
	}
}

// DiskSpaceMonitorNodeProperty configures disk space monitoring thresholds for a node
type DiskSpaceMonitorNodeProperty struct {
	XMLName                          xml.Name `xml:"hudson.node__monitors.DiskSpaceMonitorNodeProperty"`
	Class                            string   `xml:"-"`
	FreeDiskSpaceThreshold           string   `xml:"freeDiskSpaceThreshold"`
	FreeTempSpaceThreshold           string   `xml:"freeTempSpaceThreshold"`
	FreeDiskSpaceWarningThreshold    string   `xml:"freeDiskSpaceWarningThreshold,omitempty"`
	FreeTempSpaceWarningThreshold    string   `xml:"freeTempSpaceWarningThreshold,omitempty"`
}

func (d *DiskSpaceMonitorNodeProperty) GetClass() string {
	return "hudson.node_monitors.DiskSpaceMonitorNodeProperty"
}

// NewDiskSpaceMonitorNodeProperty creates a disk space monitor property
// thresholds should be strings like "1GiB", "500MiB", etc.
func NewDiskSpaceMonitorNodeProperty(freeDiskThreshold string, freeTempThreshold ...string) *DiskSpaceMonitorNodeProperty {
	prop := &DiskSpaceMonitorNodeProperty{
		Class:                      "hudson.node_monitors.DiskSpaceMonitorNodeProperty",
		FreeDiskSpaceThreshold:     freeDiskThreshold,
		FreeTempSpaceThreshold:     freeDiskThreshold,
	}
	if len(freeTempThreshold) > 0 {
		prop.FreeTempSpaceThreshold = freeTempThreshold[0]
	}
	return prop
}

// WorkspaceCleanupNodeProperty disables deferred workspace wipeout for a node
// This is from the ws-cleanup plugin
type WorkspaceCleanupNodeProperty struct {
	XMLName xml.Name `xml:"hudson.plugins.ws__cleanup.DisableDeferredWipeoutNodeProperty"`
	Class   string   `xml:"-"`
	// Plugin attribute is optional - Jenkins sets this automatically based on installed plugin version
	Plugin  string   `xml:"plugin,attr,omitempty"`
}

func (w *WorkspaceCleanupNodeProperty) GetClass() string {
	return "hudson.plugins.ws_cleanup.DisableDeferredWipeoutNodeProperty"
}

// NewWorkspaceCleanupNodeProperty creates a workspace cleanup property
// This disables deferred wipeout (requires ws-cleanup plugin)
func NewWorkspaceCleanupNodeProperty() *WorkspaceCleanupNodeProperty {
	return &WorkspaceCleanupNodeProperty{
		Class: "hudson.plugins.ws_cleanup.DisableDeferredWipeoutNodeProperty",
		// Plugin version is omitted - Jenkins will set it automatically
	}
}

// NewDeferredWipeoutNodeProperty is an alias for NewWorkspaceCleanupNodeProperty
func NewDeferredWipeoutNodeProperty() *WorkspaceCleanupNodeProperty {
	return NewWorkspaceCleanupNodeProperty()
}

// RawNodeProperty allows users to define custom node properties with arbitrary XML content
// This is useful for Jenkins plugins that aren't explicitly supported by this library.
// Users can implement their own NodeProperty types that marshal to the appropriate XML.
type RawNodeProperty struct {
	XMLName  xml.Name
	Class    string `xml:"-"`
	InnerXML string `xml:",innerxml"`
}

func (r *RawNodeProperty) GetClass() string {
	return r.Class
}

// NewRawNodeProperty creates a custom node property with the given class name and inner XML
// Example:
//   prop := NewRawNodeProperty("my.custom.NodeProperty", "<setting>value</setting>")
func NewRawNodeProperty(className string, innerXML string) *RawNodeProperty {
	return &RawNodeProperty{
		XMLName:  xml.Name{Local: className},
		Class:    className,
		InnerXML: innerXML,
	}
}

// MarshalXML handles custom marshaling for NodeProperties
func (np *NodeProperties) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "nodeProperties"
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	for i, prop := range np.Properties {
		if err := e.Encode(prop); err != nil {
			// Add context about which property failed
			return &PropertyMarshalError{
				Index:     i,
				ClassName: prop.GetClass(),
				Err:       err,
			}
		}
	}

	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// PropertyMarshalError provides context when a node property fails to marshal
type PropertyMarshalError struct {
	Index     int
	ClassName string
	Err       error
}

func (e *PropertyMarshalError) Error() string {
	return "failed to marshal node property at index " + string(rune(e.Index)) + " (" + e.ClassName + "): " + e.Err.Error()
}

func (e *PropertyMarshalError) Unwrap() error {
	return e.Err
}

// UnmarshalXML handles custom unmarshaling for NodeProperties
func (np *NodeProperties) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	np.Properties = []NodeProperty{}
	
	for {
		token, err := d.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			var prop NodeProperty
			
			// Determine the type based on the element name
			// Note: Jenkins may mangle element names by replacing . with __ in some cases
			switch t.Name.Local {
			case "hudson.slaves.EnvironmentVariablesNodeProperty":
				var envProp EnvironmentVariablesNodeProperty
				if err := d.DecodeElement(&envProp, &t); err != nil {
					return err
				}
				prop = &envProp
			case "hudson.tools.ToolLocationNodeProperty":
				var toolProp ToolLocationNodeProperty
				if err := d.DecodeElement(&toolProp, &t); err != nil {
					return err
				}
				prop = &toolProp
			case "hudson.node_monitors.DiskSpaceMonitorNodeProperty", "hudson.node__monitors.DiskSpaceMonitorNodeProperty":
				var diskProp DiskSpaceMonitorNodeProperty
				if err := d.DecodeElement(&diskProp, &t); err != nil {
					return err
				}
				prop = &diskProp
			case "hudson.slaves.WorkspaceCleanupNodeProperty", "hudson.plugins.ws__cleanup.DisableDeferredWipeoutNodeProperty":
				var workspaceProp WorkspaceCleanupNodeProperty
				if err := d.DecodeElement(&workspaceProp, &t); err != nil {
					return err
				}
				prop = &workspaceProp
			default:
				// Handle unknown/generic properties as raw XML
				var rawProp RawNodeProperty
				rawProp.XMLName = t.Name
				rawProp.Class = t.Name.Local
				if err := d.DecodeElement(&rawProp, &t); err != nil {
					return err
				}
				prop = &rawProp
			}
			
			if prop != nil {
				np.Properties = append(np.Properties, prop)
			}
			
		case xml.EndElement:
			if t.Name.Local == "nodeProperties" {
				return nil
			}
		}
	}
	
	return nil
}
