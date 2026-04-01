package types

import "time"

type File struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	IsDir    bool      `json:"is_dir"`
	MD5      string    `json:"md5,omitempty"`
	Modified time.Time `json:"modified"`
	Created  time.Time `json:"created"`
	ParentID int64     `json:"parent_id,omitempty"`
	Icon     Icon      `json:"icon,omitempty"`
}

type Icon struct {
	SmallURL  string `json:"small_url,omitempty"`
	LargeURL  string `json:"large_url,omitempty"`
	Max600    string `json:"max_600,omitempty"`
	MediumURL string `json:"medium_url,omitempty"`
}

type Folder struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	ParentID int64     `json:"parent_id"`
	Modified time.Time `json:"modified"`
	Created  time.Time `json:"created"`
}

func (f *File) GetID() string          { return f.ID }
func (f *File) GetName() string        { return f.Name }
func (f *File) GetSize() int64         { return f.Size }
func (f *File) IsDirectory() bool      { return f.IsDir }
func (f *File) GetModified() time.Time { return f.Modified }

func (f *Folder) GetID() string          { return f.ID }
func (f *Folder) GetName() string        { return f.Name }
func (f *Folder) GetSize() int64         { return 0 }
func (f *Folder) IsDirectory() bool      { return true }
func (f *Folder) GetModified() time.Time { return f.Modified }
