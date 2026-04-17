package plextrac

// Severity is the Plextrac finding severity. Plextrac expects English
// enum values on the wire even if the body is in another language.
type Severity string

const (
	SeverityCritical      Severity = "Critical"
	SeverityHigh          Severity = "High"
	SeverityMedium        Severity = "Medium"
	SeverityLow           Severity = "Low"
	SeverityInformational Severity = "Informational"
)

// Status is the finding state as stored by Plextrac.
type Status string

const (
	StatusOpen       Status = "Open"
	StatusClosed     Status = "Closed"
	StatusInProgress Status = "In Progress"
)

// Reference is a labelled URL attached to a finding.
type Reference struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// CWE is Plextrac's {id, name} shape for a CWE entry.
type CWE struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CVSSScore holds a single score entry under fields.scores.
type CVSSScore struct {
	Type        string `json:"type"`
	Label       string `json:"label"`
	Value       string `json:"value"`
	Calculation string `json:"calculation,omitempty"`
}

// Flaw is a finding in a Plextrac report.
type Flaw struct {
	ID                string         `json:"flaw_id,omitempty"`
	Title             string         `json:"title"`
	Severity          Severity       `json:"severity"`
	Status            Status         `json:"status"`
	Description       string         `json:"description"`
	Recommendations   string         `json:"recommendations"`
	References        string         `json:"references,omitempty"`
	Tags              []string       `json:"tags,omitempty"`
	CommonIdentifiers map[string]any `json:"common_identifiers,omitempty"`
	CodeSamples       []CodeSample   `json:"code_samples,omitempty"`
	Fields            map[string]any `json:"fields,omitempty"`
	CustomFields      map[string]any `json:"custom_fields,omitempty"`
}

// CodeSample is an entry in a Plextrac finding's code_samples panel.
type CodeSample struct {
	Caption  string `json:"caption"`
	Language string `json:"language,omitempty"`
	Code     string `json:"code"`
}

// ClientOrg is a Plextrac tenant client.
type ClientOrg struct {
	ID       string         `json:"client_id,omitempty"`
	Name     string         `json:"name"`
	Industry string         `json:"industry,omitempty"`
	Extra    map[string]any `json:"-"`
}

// Report is a Plextrac report.
type Report struct {
	ID        string         `json:"report_id,omitempty"`
	ClientID  string         `json:"client_id,omitempty"`
	Name      string         `json:"name"`
	Status    string         `json:"status,omitempty"`
	Template  string         `json:"template_id,omitempty"`
	CreatedAt string         `json:"created_at,omitempty"`
	UpdatedAt string         `json:"updated_at,omitempty"`
	Extra     map[string]any `json:"-"`
}

// Asset is a host/IP/URL asset that a report's findings target.
type Asset struct {
	ID          string         `json:"asset_id,omitempty"`
	Name        string         `json:"name"`
	Type        string         `json:"type,omitempty"`
	Description string         `json:"description,omitempty"`
	Extra       map[string]any `json:"-"`
}

// Writeup is a reusable finding template from the content library.
type Writeup struct {
	ID              string   `json:"writeup_id,omitempty"`
	Title           string   `json:"title"`
	Severity        Severity `json:"severity,omitempty"`
	Description     string   `json:"description,omitempty"`
	Recommendations string   `json:"recommendations,omitempty"`
	References      string   `json:"references,omitempty"`
}

// Attachment is a file uploaded to a report or finding.
type Attachment struct {
	ID       string `json:"attachment_id,omitempty"`
	Filename string `json:"filename"`
	Size     int64  `json:"size,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

// Template is a report or findings template definition.
type Template struct {
	ID    string `json:"template_id,omitempty"`
	Name  string `json:"name"`
	Kind  string `json:"kind,omitempty"`
	Owner string `json:"owner,omitempty"`
}

// Tag is a taxonomy label.
type Tag struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// User is a tenant member.
type User struct {
	ID        string `json:"user_id,omitempty"`
	Email     string `json:"email"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Role      string `json:"role,omitempty"`
}

// Export describes a requested report export.
type Export struct {
	ID     string `json:"export_id,omitempty"`
	Format string `json:"format"`
	Status string `json:"status,omitempty"`
	URL    string `json:"url,omitempty"`
}
