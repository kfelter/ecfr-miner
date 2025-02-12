package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	titlesDir = flag.String("i", "", "input directory")
	outDir    = "tmp"
)

func main() {
	flag.Parse()

	if *titlesDir == "" {
		// look for the most recent default dir
		// default dir looks like titles_1234567890

		var mostRecentDir string
		var mostRecentTime int64

		filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() && strings.HasPrefix(d.Name(), "titles_") {
				timeStr := strings.TrimPrefix(d.Name(), "titles_")
				if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
					if timestamp > mostRecentTime {
						mostRecentTime = timestamp
						mostRecentDir = d.Name()
					}
				}
			}
			return nil
		})

		if mostRecentDir != "" {
			*titlesDir = mostRecentDir
		}
	}

	// before starting to parse we should clear out any old output from previous runs
	outDir = filepath.Join(*titlesDir, "parsed")
	err := os.RemoveAll(outDir)
	if err != nil {
		log.Fatal(err)
	}

	// walk the input dir and parse the xml files into chapters and save them to sub xml files
	filepath.WalkDir(filepath.Join(*titlesDir, "raw"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), "clone.xml") {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".xml") {
			parseXML(path)
		}
		return nil
	})

}

func parseXML(path string) {
	// Read the XML file
	xmlFile, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Error reading file %s: %v", path, err)
		return
	}

	// Parse the XML file
	ecfr := &ECFR{}
	err = ecfr.Parse(xmlFile)
	if err != nil {
		log.Printf("Error parsing file %s: %v", path, err)
		log.Fatal(err)
		return
	}

	// get the title number from the file name
	titleNum := strings.TrimSuffix(filepath.Base(path), ".xml")

	// Create output directory
	outDirTitle := filepath.Join(outDir, titleNum)
	err = os.MkdirAll(outDirTitle, os.ModePerm)
	if err != nil {
		log.Printf("Error creating directory %s: %v", outDirTitle, err)
		return
	}

	// create a clone of the ecfr by marshalling and writing to file
	out, err := os.Create(filepath.Join(outDirTitle, titleNum+".clone.xml"))
	if err != nil {
		log.Printf("Error creating file %s: %v", path+".clone.xml", err)

		return

	}
	defer out.Close()

	// Marshal the ECFR struct to XML
	encoder := xml.NewEncoder(out)
	encoder.Indent("", "  ")
	err = encoder.Encode(ecfr)
	if err != nil {
		log.Printf("Error encoding file %s: %v", path+".clone.xml", err)
		return
	}

	// chapters without vols
	for _, div3 := range ecfr.DIV1.DIV3 {
		// get the chapter number
		chapterNum := div3.N
		if chapterNum == "" {
			continue
		}
		// ensure the type is "chapter"
		if div3.TYPE != "CHAPTER" {
			continue
		}

		// marshal the div to xml and write to file
		chapterFile := filepath.Join(outDirTitle, chapterNum+".xml")
		out, err := os.Create(chapterFile)
		if err != nil {
			log.Printf("Error creating file %s: %v", chapterFile, err)
			return
		}
		defer out.Close()
		encoder := xml.NewEncoder(out)
		encoder.Indent("", "  ")
		err = encoder.Encode(div3)
		if err != nil {
			log.Printf("Error encoding file %s: %v", chapterFile, err)
			return
		}
	}

	// chapters in vols
	for _, div2 := range ecfr.DIV1.DIV2 {
		if div2.TYPE != "SUBTITLE" {
			fmt.Println("Skipping [title="+titleNum+"]", "not a subtitle", div2)
			continue
		}
		err = os.MkdirAll(filepath.Join(outDirTitle, div2.N), os.ModePerm)
		if err != nil {
			log.Printf("Error creating directory %s: %v", outDirTitle, err)
			return
		}

		for _, div3 := range div2.DIV3 {
			// get the chapter number
			chapterNum := div3.N
			if chapterNum == "" {
				continue
			}
			// ensure the type is "chapter"
			if div3.TYPE != "CHAPTER" {
				continue
			}

			// marshal the div to xml and write to file
			chapterFile := filepath.Join(outDirTitle, chapterNum+".xml")
			out, err := os.Create(chapterFile)
			if err != nil {
				log.Printf("Error creating file %s: %v", chapterFile, err)
				return
			}
			defer out.Close()
			encoder := xml.NewEncoder(out)
			encoder.Indent("", "  ")
			err = encoder.Encode(div3)
			if err != nil {
				log.Printf("Error encoding file %s: %v", chapterFile, err)
				return
			}

			chapterFileSub := filepath.Join(outDirTitle, div2.N, chapterNum+".xml")
			out, err = os.Create(chapterFileSub)
			if err != nil {
				log.Printf("Error creating file %s: %v", chapterFileSub, err)
				return
			}
			defer out.Close()
			encoder = xml.NewEncoder(out)
			encoder.Indent("", "  ")
			err = encoder.Encode(div3)
			if err != nil {
				log.Printf("Error encoding file %s: %v", chapterFileSub, err)
				return
			}

		}
	}

}

type ECFR struct {
	XMLName xml.Name `xml:"ECFR" json:"ecfr,omitempty"`
	Text    string   `xml:",chardata" json:"text,omitempty"`
	AMDDATE string   `xml:"AMDDATE"`
	VOLUME  struct {
		Text    string `xml:",chardata" json:"text,omitempty"`
		N       string `xml:"N,attr" json:"n,omitempty"`
		AMDDATE string `xml:"AMDDATE,attr" json:"amddate,omitempty"`
	} `xml:"VOLUME" json:"volume,omitempty"`
	DIV1 struct {
		Text string `xml:",chardata" json:"text,omitempty"`
		N    string `xml:"N,attr" json:"n,omitempty"`
		TYPE string `xml:"TYPE,attr" json:"type,omitempty"`
		HEAD string `xml:"HEAD"`
		DIV3 []DIV3 `xml:"DIV3" json:"div3,omitempty"`
		DIV2 []struct {
			Text string `xml:",chardata" json:"text,omitempty"`
			N    string `xml:"N,attr" json:"n,omitempty"`
			TYPE string `xml:"TYPE,attr" json:"type,omitempty"`
			HEAD string `xml:"HEAD"`
			DIV5 struct {
				Text   string `xml:",chardata" json:"text,omitempty"`
				N      string `xml:"N,attr" json:"n,omitempty"`
				TYPE   string `xml:"TYPE,attr" json:"type,omitempty"`
				VOLUME string `xml:"VOLUME,attr" json:"volume,omitempty"`
				HEAD   string `xml:"HEAD"`
				AUTH   struct {
					Text   string `xml:",chardata" json:"text,omitempty"`
					HED    string `xml:"HED"`
					PSPACE string `xml:"PSPACE"`
				} `xml:"AUTH" json:"auth,omitempty"`
				SOURCE struct {
					Text   string `xml:",chardata" json:"text,omitempty"`
					HED    string `xml:"HED"`
					PSPACE string `xml:"PSPACE"`
				} `xml:"SOURCE" json:"source,omitempty"`
				DIV6 []struct {
					Text   string `xml:",chardata" json:"text,omitempty"`
					N      string `xml:"N,attr" json:"n,omitempty"`
					TYPE   string `xml:"TYPE,attr" json:"type,omitempty"`
					HEAD   string `xml:"HEAD"`
					EDNOTE struct {
						Text   string `xml:",chardata" json:"text,omitempty"`
						HED    string `xml:"HED"`
						PSPACE string `xml:"PSPACE"`
					} `xml:"EDNOTE" json:"ednote,omitempty"`
					DIV8 []struct {
						Text   string `xml:",chardata" json:"text,omitempty"`
						N      string `xml:"N,attr" json:"n,omitempty"`
						TYPE   string `xml:"TYPE,attr" json:"type,omitempty"`
						VOLUME string `xml:"VOLUME,attr" json:"volume,omitempty"`
						HEAD   string `xml:"HEAD"`
						P      []struct {
							Text string `xml:",chardata" json:"text,omitempty"`
							I    string `xml:"I"`
							E    struct {
								Text string `xml:",chardata" json:"text,omitempty"`
								T    string `xml:"T,attr" json:"t,omitempty"`
							} `xml:"E" json:"e,omitempty"`
						} `xml:"P" json:"p,omitempty"`
					} `xml:"DIV8" json:"div8,omitempty"`
				} `xml:"DIV6" json:"div6,omitempty"`
			} `xml:"DIV5" json:"div5,omitempty"`

			DIV3 []DIV3 `xml:"DIV3" json:"div3,omitempty"`
		} `xml:"DIV2" json:"div2,omitempty"`
	} `xml:"DIV1" json:"div1,omitempty"`
}

type DIV3 struct {
	Text string `xml:",chardata" json:"text,omitempty"`
	N    string `xml:"N,attr" json:"n,omitempty"`
	TYPE string `xml:"TYPE,attr" json:"type,omitempty"`
	HEAD string `xml:"HEAD"`
	DIV5 []struct {
		Text   string `xml:",chardata" json:"text,omitempty"`
		N      string `xml:"N,attr" json:"n,omitempty"`
		TYPE   string `xml:"TYPE,attr" json:"type,omitempty"`
		VOLUME string `xml:"VOLUME,attr" json:"volume,omitempty"`
		HEAD   string `xml:"HEAD"`
		AUTH   struct {
			Text   string `xml:",chardata" json:"text,omitempty"`
			HED    string `xml:"HED"`
			PSPACE struct {
				Text string   `xml:",chardata" json:"text,omitempty"`
				I    []string `xml:"I"`
			} `xml:"PSPACE" json:"pspace,omitempty"`
		} `xml:"AUTH" json:"auth,omitempty"`
		SOURCE struct {
			Text   string `xml:",chardata" json:"text,omitempty"`
			HED    string `xml:"HED"`
			PSPACE string `xml:"PSPACE"`
		} `xml:"SOURCE" json:"source,omitempty"`
		DIV6 []struct {
			Text string `xml:",chardata" json:"text,omitempty"`
			N    string `xml:"N,attr" json:"n,omitempty"`
			TYPE string `xml:"TYPE,attr" json:"type,omitempty"`
			HEAD string `xml:"HEAD"`
			DIV8 []struct {
				Text   string `xml:",chardata" json:"text,omitempty"`
				N      string `xml:"N,attr" json:"n,omitempty"`
				TYPE   string `xml:"TYPE,attr" json:"type,omitempty"`
				VOLUME string `xml:"VOLUME,attr" json:"volume,omitempty"`
				HEAD   struct {
					Text string `xml:",chardata" json:"text,omitempty"`
					E    struct {
						Text string `xml:",chardata" json:"text,omitempty"`
						T    string `xml:"T,attr" json:"t,omitempty"`
					} `xml:"E" json:"e,omitempty"`
				} `xml:"HEAD" json:"head,omitempty"`
				P []struct {
					Text string   `xml:",chardata" json:"text,omitempty"`
					I    []string `xml:"I"`
					E    []struct {
						Text string `xml:",chardata" json:"text,omitempty"`
						T    string `xml:"T,attr" json:"t,omitempty"`
					} `xml:"E" json:"e,omitempty"`
				} `xml:"P" json:"p,omitempty"`
				CITA struct {
					Text string `xml:",chardata" json:"text,omitempty"`
					TYPE string `xml:"TYPE,attr" json:"type,omitempty"`
				} `xml:"CITA" json:"cita,omitempty"`
				EXTRACT []struct {
					Text   string `xml:",chardata" json:"text,omitempty"`
					FPDASH string `xml:"FP-DASH"`
					P      []struct {
						Text string `xml:",chardata" json:"text,omitempty"`
						I    string `xml:"I"`
					} `xml:"P" json:"p,omitempty"`
				} `xml:"EXTRACT" json:"extract,omitempty"`
				DIV []struct {
					Text  string `xml:",chardata" json:"text,omitempty"`
					Width string `xml:"width,attr" json:"width,omitempty"`
					DIV   struct {
						Text  string `xml:",chardata" json:"text,omitempty"`
						Class string `xml:"class,attr" json:"class,omitempty"`
						TABLE struct {
							Text        string `xml:",chardata" json:"text,omitempty"`
							Border      string `xml:"border,attr" json:"border,omitempty"`
							Cellpadding string `xml:"cellpadding,attr" json:"cellpadding,omitempty"`
							Cellspacing string `xml:"cellspacing,attr" json:"cellspacing,omitempty"`
							Class       string `xml:"class,attr" json:"class,omitempty"`
							Frame       string `xml:"frame,attr" json:"frame,omitempty"`
							Width       string `xml:"width,attr" json:"width,omitempty"`
							CAPTION     struct {
								Text string `xml:",chardata" json:"text,omitempty"`
								P    struct {
									Text  string `xml:",chardata" json:"text,omitempty"`
									Class string `xml:"class,attr" json:"class,omitempty"`
									E     struct {
										Text string `xml:",chardata" json:"text,omitempty"`
										T    string `xml:"T,attr" json:"t,omitempty"`
									} `xml:"E" json:"e,omitempty"`
								} `xml:"P" json:"p,omitempty"`
							} `xml:"CAPTION" json:"caption,omitempty"`
							THEAD struct {
								Text string `xml:",chardata" json:"text,omitempty"`
								TR   struct {
									Text string `xml:",chardata" json:"text,omitempty"`
									TH   []struct {
										Text  string `xml:",chardata" json:"text,omitempty"`
										Class string `xml:"class,attr" json:"class,omitempty"`
										Br    string `xml:"br"`
										E     struct {
											Text string `xml:",chardata" json:"text,omitempty"`
											T    string `xml:"T,attr" json:"t,omitempty"`
										} `xml:"E" json:"e,omitempty"`
									} `xml:"TH" json:"th,omitempty"`
								} `xml:"TR" json:"tr,omitempty"`
							} `xml:"THEAD" json:"thead,omitempty"`
							TBODY struct {
								Text string `xml:",chardata" json:"text,omitempty"`
								TR   []struct {
									Text string `xml:",chardata" json:"text,omitempty"`
									TD   []struct {
										Text  string `xml:",chardata" json:"text,omitempty"`
										Class string `xml:"class,attr" json:"class,omitempty"`
										E     struct {
											Text string `xml:",chardata" json:"text,omitempty"`
											T    string `xml:"T,attr" json:"t,omitempty"`
										} `xml:"E" json:"e,omitempty"`
									} `xml:"TD" json:"td,omitempty"`
								} `xml:"TR" json:"tr,omitempty"`
							} `xml:"TBODY" json:"tbody,omitempty"`
							TFOOT struct {
								Text string `xml:",chardata" json:"text,omitempty"`
								TR   []struct {
									Text string `xml:",chardata" json:"text,omitempty"`
									TD   struct {
										Text    string `xml:",chardata" json:"text,omitempty"`
										Colspan string `xml:"colspan,attr" json:"colspan,omitempty"`
										E       struct {
											Text string `xml:",chardata" json:"text,omitempty"`
											T    string `xml:"T,attr" json:"t,omitempty"`
										} `xml:"E" json:"e,omitempty"`
									} `xml:"TD" json:"td,omitempty"`
								} `xml:"TR" json:"tr,omitempty"`
							} `xml:"TFOOT" json:"tfoot,omitempty"`
						} `xml:"TABLE" json:"table,omitempty"`
					} `xml:"DIV" json:"div,omitempty"`
				} `xml:"DIV" json:"div,omitempty"`
				FP1     []string `xml:"FP-1"`
				FP      string   `xml:"FP"`
				SECAUTH struct {
					Text string `xml:",chardata" json:"text,omitempty"`
					TYPE string `xml:"TYPE,attr" json:"type,omitempty"`
				} `xml:"SECAUTH" json:"secauth,omitempty"`
			} `xml:"DIV8" json:"div8,omitempty"`
			DIV9 struct {
				Text   string `xml:",chardata" json:"text,omitempty"`
				N      string `xml:"N,attr" json:"n,omitempty"`
				TYPE   string `xml:"TYPE,attr" json:"type,omitempty"`
				VOLUME string `xml:"VOLUME,attr" json:"volume,omitempty"`
				HEAD   string `xml:"HEAD"`
				DIV    struct {
					Text  string `xml:",chardata" json:"text,omitempty"`
					Width string `xml:"width,attr" json:"width,omitempty"`
					DIV   struct {
						Text  string `xml:",chardata" json:"text,omitempty"`
						Class string `xml:"class,attr" json:"class,omitempty"`
						TABLE struct {
							Text        string `xml:",chardata" json:"text,omitempty"`
							Border      string `xml:"border,attr" json:"border,omitempty"`
							Cellpadding string `xml:"cellpadding,attr" json:"cellpadding,omitempty"`
							Cellspacing string `xml:"cellspacing,attr" json:"cellspacing,omitempty"`
							Class       string `xml:"class,attr" json:"class,omitempty"`
							Frame       string `xml:"frame,attr" json:"frame,omitempty"`
							Width       string `xml:"width,attr" json:"width,omitempty"`
							THEAD       struct {
								Text string `xml:",chardata" json:"text,omitempty"`
								TR   struct {
									Text string `xml:",chardata" json:"text,omitempty"`
									TH   []struct {
										Text  string `xml:",chardata" json:"text,omitempty"`
										Class string `xml:"class,attr" json:"class,omitempty"`
									} `xml:"TH" json:"th,omitempty"`
								} `xml:"TR" json:"tr,omitempty"`
							} `xml:"THEAD" json:"thead,omitempty"`
							TBODY struct {
								Text string `xml:",chardata" json:"text,omitempty"`
								TR   []struct {
									Text string `xml:",chardata" json:"text,omitempty"`
									TD   []struct {
										Text  string   `xml:",chardata" json:"text,omitempty"`
										Class string   `xml:"class,attr" json:"class,omitempty"`
										Br    []string `xml:"br"`
										E     struct {
											Text string `xml:",chardata" json:"text,omitempty"`
											T    string `xml:"T,attr" json:"t,omitempty"`
										} `xml:"E" json:"e,omitempty"`
									} `xml:"TD" json:"td,omitempty"`
								} `xml:"TR" json:"tr,omitempty"`
							} `xml:"TBODY" json:"tbody,omitempty"`
						} `xml:"TABLE" json:"table,omitempty"`
					} `xml:"DIV" json:"div,omitempty"`
				} `xml:"DIV" json:"div,omitempty"`
				P []struct {
					Text string   `xml:",chardata" json:"text,omitempty"`
					I    []string `xml:"I"`
				} `xml:"P" json:"p,omitempty"`
				CITA struct {
					Text string `xml:",chardata" json:"text,omitempty"`
					TYPE string `xml:"TYPE,attr" json:"type,omitempty"`
				} `xml:"CITA" json:"cita,omitempty"`
				FP1 []string `xml:"FP-1"`
				HD3 []string `xml:"HD3"`
			} `xml:"DIV9" json:"div9,omitempty"`
			DIV7 []struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				N    string `xml:"N,attr" json:"n,omitempty"`
				TYPE string `xml:"TYPE,attr" json:"type,omitempty"`
				HEAD string `xml:"HEAD"`
				DIV8 []struct {
					Text   string `xml:",chardata" json:"text,omitempty"`
					N      string `xml:"N,attr" json:"n,omitempty"`
					TYPE   string `xml:"TYPE,attr" json:"type,omitempty"`
					VOLUME string `xml:"VOLUME,attr" json:"volume,omitempty"`
					HEAD   string `xml:"HEAD"`
					P      []struct {
						Text string   `xml:",chardata" json:"text,omitempty"`
						I    []string `xml:"I"`
					} `xml:"P" json:"p,omitempty"`
					CITA struct {
						Text string `xml:",chardata" json:"text,omitempty"`
						TYPE string `xml:"TYPE,attr" json:"type,omitempty"`
					} `xml:"CITA" json:"cita,omitempty"`
					HD1 string `xml:"HD1"`
					DIV struct {
						Text  string `xml:",chardata" json:"text,omitempty"`
						Width string `xml:"width,attr" json:"width,omitempty"`
						DIV   struct {
							Text  string `xml:",chardata" json:"text,omitempty"`
							Class string `xml:"class,attr" json:"class,omitempty"`
							TABLE struct {
								Text        string `xml:",chardata" json:"text,omitempty"`
								Border      string `xml:"border,attr" json:"border,omitempty"`
								Cellpadding string `xml:"cellpadding,attr" json:"cellpadding,omitempty"`
								Cellspacing string `xml:"cellspacing,attr" json:"cellspacing,omitempty"`
								Class       string `xml:"class,attr" json:"class,omitempty"`
								Frame       string `xml:"frame,attr" json:"frame,omitempty"`
								Width       string `xml:"width,attr" json:"width,omitempty"`
								CAPTION     struct {
									Text string `xml:",chardata" json:"text,omitempty"`
									P    struct {
										Text  string `xml:",chardata" json:"text,omitempty"`
										Class string `xml:"class,attr" json:"class,omitempty"`
										E     struct {
											Text string `xml:",chardata" json:"text,omitempty"`
											T    string `xml:"T,attr" json:"t,omitempty"`
										} `xml:"E" json:"e,omitempty"`
									} `xml:"P" json:"p,omitempty"`
								} `xml:"CAPTION" json:"caption,omitempty"`
								THEAD struct {
									Text string `xml:",chardata" json:"text,omitempty"`
									TR   struct {
										Text string `xml:",chardata" json:"text,omitempty"`
										TH   []struct {
											Text  string `xml:",chardata" json:"text,omitempty"`
											Class string `xml:"class,attr" json:"class,omitempty"`
										} `xml:"TH" json:"th,omitempty"`
									} `xml:"TR" json:"tr,omitempty"`
								} `xml:"THEAD" json:"thead,omitempty"`
								TBODY struct {
									Text string `xml:",chardata" json:"text,omitempty"`
									TR   []struct {
										Text string `xml:",chardata" json:"text,omitempty"`
										TD   []struct {
											Text  string `xml:",chardata" json:"text,omitempty"`
											Class string `xml:"class,attr" json:"class,omitempty"`
										} `xml:"TD" json:"td,omitempty"`
									} `xml:"TR" json:"tr,omitempty"`
								} `xml:"TBODY" json:"tbody,omitempty"`
							} `xml:"TABLE" json:"table,omitempty"`
						} `xml:"DIV" json:"div,omitempty"`
					} `xml:"DIV" json:"div,omitempty"`
				} `xml:"DIV8" json:"div8,omitempty"`
			} `xml:"DIV7" json:"div7,omitempty"`
			SOURCE struct {
				Text   string `xml:",chardata" json:"text,omitempty"`
				HED    string `xml:"HED"`
				PSPACE string `xml:"PSPACE"`
			} `xml:"SOURCE" json:"source,omitempty"`
		} `xml:"DIV6" json:"div6,omitempty"`
		DIV9 []struct {
			Text   string   `xml:",chardata" json:"text,omitempty"`
			N      string   `xml:"N,attr" json:"n,omitempty"`
			TYPE   string   `xml:"TYPE,attr" json:"type,omitempty"`
			VOLUME string   `xml:"VOLUME,attr" json:"volume,omitempty"`
			HEAD   string   `xml:"HEAD"`
			HD1    []string `xml:"HD1"`
			P      []struct {
				Text string   `xml:",chardata" json:"text,omitempty"`
				I    []string `xml:"I"`
			} `xml:"P" json:"p,omitempty"`
			CITA struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				TYPE string `xml:"TYPE,attr" json:"type,omitempty"`
			} `xml:"CITA" json:"cita,omitempty"`
			Img struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				Src  string `xml:"src,attr" json:"src,omitempty"`
			} `xml:"img" json:"img,omitempty"`
			FP  []string `xml:"FP"`
			HD2 []struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				I    string `xml:"I"`
			} `xml:"HD2" json:"hd2,omitempty"`
			HD3    []string `xml:"HD3"`
			FPDASH []string `xml:"FP-DASH"`
			FP1    []string `xml:"FP-1"`
			XREF   struct {
				Text  string `xml:",chardata" json:"text,omitempty"`
				ID    string `xml:"ID,attr" json:"id,omitempty"`
				REFID string `xml:"REFID,attr" json:"refid,omitempty"`
			} `xml:"XREF" json:"xref,omitempty"`
			NOTE struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				HED  string `xml:"HED"`
				P    string `xml:"P"`
			} `xml:"NOTE" json:"note,omitempty"`
		} `xml:"DIV9" json:"div9,omitempty"`
		DIV8 []struct {
			Text   string `xml:",chardata" json:"text,omitempty"`
			N      string `xml:"N,attr" json:"n,omitempty"`
			TYPE   string `xml:"TYPE,attr" json:"type,omitempty"`
			VOLUME string `xml:"VOLUME,attr" json:"volume,omitempty"`
			HEAD   struct {
				Text  string `xml:",chardata" json:"text,omitempty"`
				SU    string `xml:"SU"`
				FTREF string `xml:"FTREF"`
			} `xml:"HEAD" json:"head,omitempty"`
			P []struct {
				Text string   `xml:",chardata" json:"text,omitempty"`
				I    []string `xml:"I"`
				E    struct {
					Text string `xml:",chardata" json:"text,omitempty"`
					T    string `xml:"T,attr" json:"t,omitempty"`
				} `xml:"E" json:"e,omitempty"`
			} `xml:"P" json:"p,omitempty"`
			CITA struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				TYPE string `xml:"TYPE,attr" json:"type,omitempty"`
			} `xml:"CITA" json:"cita,omitempty"`
			DIV struct {
				Text  string `xml:",chardata" json:"text,omitempty"`
				Width string `xml:"width,attr" json:"width,omitempty"`
				DIV   struct {
					Text  string `xml:",chardata" json:"text,omitempty"`
					Class string `xml:"class,attr" json:"class,omitempty"`
					TABLE struct {
						Text        string `xml:",chardata" json:"text,omitempty"`
						Border      string `xml:"border,attr" json:"border,omitempty"`
						Cellpadding string `xml:"cellpadding,attr" json:"cellpadding,omitempty"`
						Cellspacing string `xml:"cellspacing,attr" json:"cellspacing,omitempty"`
						Class       string `xml:"class,attr" json:"class,omitempty"`
						Frame       string `xml:"frame,attr" json:"frame,omitempty"`
						Width       string `xml:"width,attr" json:"width,omitempty"`
						THEAD       struct {
							Text string `xml:",chardata" json:"text,omitempty"`
							TR   struct {
								Text string `xml:",chardata" json:"text,omitempty"`
								TH   []struct {
									Text  string   `xml:",chardata" json:"text,omitempty"`
									Class string   `xml:"class,attr" json:"class,omitempty"`
									Br    []string `xml:"br"`
								} `xml:"TH" json:"th,omitempty"`
							} `xml:"TR" json:"tr,omitempty"`
						} `xml:"THEAD" json:"thead,omitempty"`
						TBODY struct {
							Text string `xml:",chardata" json:"text,omitempty"`
							TR   []struct {
								Text string `xml:",chardata" json:"text,omitempty"`
								TD   []struct {
									Text  string `xml:",chardata" json:"text,omitempty"`
									Class string `xml:"class,attr" json:"class,omitempty"`
								} `xml:"TD" json:"td,omitempty"`
							} `xml:"TR" json:"tr,omitempty"`
						} `xml:"TBODY" json:"tbody,omitempty"`
					} `xml:"TABLE" json:"table,omitempty"`
				} `xml:"DIV" json:"div,omitempty"`
			} `xml:"DIV" json:"div,omitempty"`
			SECAUTH struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				TYPE string `xml:"TYPE,attr" json:"type,omitempty"`
			} `xml:"SECAUTH" json:"secauth,omitempty"`
			FTNT struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				P    struct {
					Text string `xml:",chardata" json:"text,omitempty"`
					SU   string `xml:"SU"`
				} `xml:"P" json:"p,omitempty"`
			} `xml:"FTNT" json:"ftnt,omitempty"`
			XREF struct {
				Text  string `xml:",chardata" json:"text,omitempty"`
				ID    string `xml:"ID,attr" json:"id,omitempty"`
				REFID string `xml:"REFID,attr" json:"refid,omitempty"`
			} `xml:"XREF" json:"xref,omitempty"`
			AUTH struct {
				Text   string `xml:",chardata" json:"text,omitempty"`
				HED    string `xml:"HED"`
				PSPACE string `xml:"PSPACE"`
			} `xml:"AUTH" json:"auth,omitempty"`
		} `xml:"DIV8" json:"div8,omitempty"`
		XREF []struct {
			Text  string `xml:",chardata" json:"text,omitempty"`
			ID    string `xml:"ID,attr" json:"id,omitempty"`
			REFID string `xml:"REFID,attr" json:"refid,omitempty"`
		} `xml:"XREF" json:"xref,omitempty"`
	} `xml:"DIV5" json:"div5,omitempty"`
	DIV4 []struct {
		Text   string `xml:",chardata" json:"text,omitempty"`
		N      string `xml:"N,attr" json:"n,omitempty"`
		TYPE   string `xml:"TYPE,attr" json:"type,omitempty"`
		HEAD   string `xml:"HEAD"`
		SOURCE struct {
			Text   string `xml:",chardata" json:"text,omitempty"`
			HED    string `xml:"HED"`
			PSPACE string `xml:"PSPACE"`
		} `xml:"SOURCE" json:"source,omitempty"`
		DIV5 []struct {
			Text   string `xml:",chardata" json:"text,omitempty"`
			N      string `xml:"N,attr" json:"n,omitempty"`
			TYPE   string `xml:"TYPE,attr" json:"type,omitempty"`
			VOLUME string `xml:"VOLUME,attr" json:"volume,omitempty"`
			HEAD   string `xml:"HEAD"`
			AUTH   struct {
				Text   string `xml:",chardata" json:"text,omitempty"`
				HED    string `xml:"HED"`
				PSPACE string `xml:"PSPACE"`
			} `xml:"AUTH" json:"auth,omitempty"`
			SOURCE struct {
				Text   string `xml:",chardata" json:"text,omitempty"`
				HED    string `xml:"HED"`
				PSPACE string `xml:"PSPACE"`
			} `xml:"SOURCE" json:"source,omitempty"`
			DIV8 []struct {
				Text   string `xml:",chardata" json:"text,omitempty"`
				N      string `xml:"N,attr" json:"n,omitempty"`
				TYPE   string `xml:"TYPE,attr" json:"type,omitempty"`
				VOLUME string `xml:"VOLUME,attr" json:"volume,omitempty"`
				HEAD   string `xml:"HEAD"`
				P      []struct {
					Text string `xml:",chardata" json:"text,omitempty"`
					I    string `xml:"I"`
				} `xml:"P" json:"p,omitempty"`
				DIV struct {
					Text  string `xml:",chardata" json:"text,omitempty"`
					Width string `xml:"width,attr" json:"width,omitempty"`
					DIV   struct {
						Text  string `xml:",chardata" json:"text,omitempty"`
						Class string `xml:"class,attr" json:"class,omitempty"`
						TABLE struct {
							Text        string `xml:",chardata" json:"text,omitempty"`
							Border      string `xml:"border,attr" json:"border,omitempty"`
							Cellpadding string `xml:"cellpadding,attr" json:"cellpadding,omitempty"`
							Cellspacing string `xml:"cellspacing,attr" json:"cellspacing,omitempty"`
							Class       string `xml:"class,attr" json:"class,omitempty"`
							Frame       string `xml:"frame,attr" json:"frame,omitempty"`
							Width       string `xml:"width,attr" json:"width,omitempty"`
							CAPTION     struct {
								Text string `xml:",chardata" json:"text,omitempty"`
								P    struct {
									Text  string `xml:",chardata" json:"text,omitempty"`
									Class string `xml:"class,attr" json:"class,omitempty"`
									E     struct {
										Text string `xml:",chardata" json:"text,omitempty"`
										T    string `xml:"T,attr" json:"t,omitempty"`
									} `xml:"E" json:"e,omitempty"`
								} `xml:"P" json:"p,omitempty"`
							} `xml:"CAPTION" json:"caption,omitempty"`
							THEAD struct {
								Text string `xml:",chardata" json:"text,omitempty"`
								TR   struct {
									Text string `xml:",chardata" json:"text,omitempty"`
									TH   []struct {
										Text  string `xml:",chardata" json:"text,omitempty"`
										Class string `xml:"class,attr" json:"class,omitempty"`
									} `xml:"TH" json:"th,omitempty"`
								} `xml:"TR" json:"tr,omitempty"`
							} `xml:"THEAD" json:"thead,omitempty"`
							TBODY struct {
								Text string `xml:",chardata" json:"text,omitempty"`
								TR   []struct {
									Text string `xml:",chardata" json:"text,omitempty"`
									TD   []struct {
										Text  string `xml:",chardata" json:"text,omitempty"`
										Class string `xml:"class,attr" json:"class,omitempty"`
									} `xml:"TD" json:"td,omitempty"`
								} `xml:"TR" json:"tr,omitempty"`
							} `xml:"TBODY" json:"tbody,omitempty"`
						} `xml:"TABLE" json:"table,omitempty"`
					} `xml:"DIV" json:"div,omitempty"`
				} `xml:"DIV" json:"div,omitempty"`
			} `xml:"DIV8" json:"div8,omitempty"`
			DIV6 []struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				N    string `xml:"N,attr" json:"n,omitempty"`
				TYPE string `xml:"TYPE,attr" json:"type,omitempty"`
				HEAD string `xml:"HEAD"`
				DIV8 []struct {
					Text   string `xml:",chardata" json:"text,omitempty"`
					N      string `xml:"N,attr" json:"n,omitempty"`
					TYPE   string `xml:"TYPE,attr" json:"type,omitempty"`
					VOLUME string `xml:"VOLUME,attr" json:"volume,omitempty"`
					HEAD   string `xml:"HEAD"`
					P      []struct {
						Text string   `xml:",chardata" json:"text,omitempty"`
						I    []string `xml:"I"`
						E    struct {
							Text string `xml:",chardata" json:"text,omitempty"`
							T    string `xml:"T,attr" json:"t,omitempty"`
						} `xml:"E" json:"e,omitempty"`
					} `xml:"P" json:"p,omitempty"`
					DIV struct {
						Text  string `xml:",chardata" json:"text,omitempty"`
						Width string `xml:"width,attr" json:"width,omitempty"`
						DIV   struct {
							Text  string `xml:",chardata" json:"text,omitempty"`
							Class string `xml:"class,attr" json:"class,omitempty"`
							TABLE struct {
								Text        string `xml:",chardata" json:"text,omitempty"`
								Border      string `xml:"border,attr" json:"border,omitempty"`
								Cellpadding string `xml:"cellpadding,attr" json:"cellpadding,omitempty"`
								Cellspacing string `xml:"cellspacing,attr" json:"cellspacing,omitempty"`
								Class       string `xml:"class,attr" json:"class,omitempty"`
								Frame       string `xml:"frame,attr" json:"frame,omitempty"`
								Width       string `xml:"width,attr" json:"width,omitempty"`
								THEAD       struct {
									Text string `xml:",chardata" json:"text,omitempty"`
									TR   struct {
										Text string `xml:",chardata" json:"text,omitempty"`
										TH   []struct {
											Text  string   `xml:",chardata" json:"text,omitempty"`
											Class string   `xml:"class,attr" json:"class,omitempty"`
											Br    []string `xml:"br"`
										} `xml:"TH" json:"th,omitempty"`
									} `xml:"TR" json:"tr,omitempty"`
								} `xml:"THEAD" json:"thead,omitempty"`
								TBODY struct {
									Text string `xml:",chardata" json:"text,omitempty"`
									TR   []struct {
										Text string `xml:",chardata" json:"text,omitempty"`
										TD   []struct {
											Text  string `xml:",chardata" json:"text,omitempty"`
											Class string `xml:"class,attr" json:"class,omitempty"`
										} `xml:"TD" json:"td,omitempty"`
									} `xml:"TR" json:"tr,omitempty"`
								} `xml:"TBODY" json:"tbody,omitempty"`
							} `xml:"TABLE" json:"table,omitempty"`
						} `xml:"DIV" json:"div,omitempty"`
					} `xml:"DIV" json:"div,omitempty"`
				} `xml:"DIV8" json:"div8,omitempty"`
			} `xml:"DIV6" json:"div6,omitempty"`
			DIV9 []struct {
				Text   string   `xml:",chardata" json:"text,omitempty"`
				N      string   `xml:"N,attr" json:"n,omitempty"`
				TYPE   string   `xml:"TYPE,attr" json:"type,omitempty"`
				VOLUME string   `xml:"VOLUME,attr" json:"volume,omitempty"`
				HEAD   string   `xml:"HEAD"`
				HD1    []string `xml:"HD1"`
				P      []struct {
					Text string   `xml:",chardata" json:"text,omitempty"`
					I    []string `xml:"I"`
					E    struct {
						Text string `xml:",chardata" json:"text,omitempty"`
						T    string `xml:"T,attr" json:"t,omitempty"`
					} `xml:"E" json:"e,omitempty"`
				} `xml:"P" json:"p,omitempty"`
			} `xml:"DIV9" json:"div9,omitempty"`
		} `xml:"DIV5" json:"div5,omitempty"`
	} `xml:"DIV4" json:"div4,omitempty"`
}

func (ecfr *ECFR) Parse(data []byte) error {
	return xml.Unmarshal(data, ecfr)
}
