package binfiles

import (
	"embed"
	"gopkg.in/macaron.v1"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

type BinFiles struct {
	fs         *embed.FS
	baseFolder string
}

type templateFile struct {
	key  string
	name string
	ext  string
}

// New Create a new BinFiles object
//  fs = embedded filesystem
//  baseFolder= base folder of embedded filesystem. It will be stripped off from filenames
func New(fs *embed.FS, baseFolder string) BinFiles {
	if !strings.HasSuffix(baseFolder, "/") {
		baseFolder = baseFolder + "/"
	}

	return BinFiles{
		fs:         fs,
		baseFolder: baseFolder,
	}
}

// Open is used to serve static files
func (b BinFiles) Open(name string) (http.File, error) {

	var fname string
	fname = path.Join(b.baseFolder, name)

	//log.Println("Opening static file " + fname)
	file, err := b.fs.Open(fname)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer file.Close()
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	hf := NewHttpFile(fname, content, defaultFileTimestamp)
	return hf, err
}

func (b BinFiles) ListFiles() (filelist []macaron.TemplateFile) {

	list, err := getAllFilenames(b.fs, ".")
	if err != nil {
		log.Println("Error reading file list in embedded filesystem: %s", err)
	}

	//log.Println("File list:")
	for _, filename := range list {

		// remove base folder
		if strings.HasPrefix(filename, b.baseFolder) {
			filename = filename[len(b.baseFolder):]
		}

		//log.Println(filename)
		ext := macaron.GetExt(filename)
		filelist = append(filelist, templateFile{
			key:  filename,
			name: filename[0 : len(filename)-len(ext)],
			ext:  ext,
		})
	}

	return
}

func (b BinFiles) Get(s string) (io.Reader, error) {
	log.Println("Get file " + s)
	return b.fs.Open(b.baseFolder + s)

}

func (tf templateFile) Name() string {
	//log.Println("Name: " + tf.name)
	return tf.name
}

func (tf templateFile) Ext() string {
	//log.Println("Ext: " + tf.ext)
	return tf.ext
}

func (tf templateFile) Data() []byte {
	log.Println("Get Data: " + tf.key + " (not implemented yet)")
	return []byte{}
}

func getAllFilenames(fs *embed.FS, path string) (out []string, err error) {
	if len(path) == 0 {
		path = "."
	}
	entries, err := fs.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		fp := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			res, err := getAllFilenames(fs, fp)
			if err != nil {
				return nil, err
			}
			out = append(out, res...)
			continue
		}
		out = append(out, fp)
	}
	return
}
