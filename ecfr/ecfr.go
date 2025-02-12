package ecfr

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

func (ecfr *ECFR) CountWordsChapter(chapter string) (int, error) {
	if ecfr == nil {
		return 0, fmt.Errorf("ecfr is nil")
	}
	// chapters in vols
	for _, div2 := range ecfr.DIV1.DIV2 {
		if div2.TYPE != "SUBTITLE" {
			fmt.Println("Skipping", "not a subtitle", div2)
			continue
		}

		// chapters in subtitles
		for _, div3 := range div2.DIV3 {
			// get the chapter number
			chapterNum := div3.N
			if chapterNum != chapter {
				continue
			}
			// ensure the type is "chapter"
			if div3.TYPE != "CHAPTER" {
				continue
			}

			// count the words in the chapter
			buf := new(bytes.Buffer)
			enc := xml.NewEncoder(buf)
			enc.Indent("", "")
			err := enc.Encode(div3)
			if err != nil {
				log.Fatal(err)
			}

			return CountWords(bufio.NewReader(buf))

		}
	}

	// chapters without vols
	for _, div3 := range ecfr.DIV1.DIV3 {
		// get the chapter number
		chapterNum := div3.N
		if chapterNum != chapter {
			continue
		}
		// ensure the type is "chapter"
		if div3.TYPE != "CHAPTER" {
			continue
		}

		// count the words in the chapter
		buf := new(bytes.Buffer)
		enc := xml.NewEncoder(buf)
		enc.Indent("", "")
		err := enc.Encode(div3)
		if err != nil {
			log.Fatal(err)
		}
		return CountWords(bufio.NewReader(buf))
	}

	fmt.Println("title", ecfr.DIV1.N, "chapter", chapter, "not found")
	return 0, fmt.Errorf("chapter not found")
}

func (ecfr *ECFR) CountWordsSubtitle(subtitle string) (int, error) {
	// count all the chapters in a subtitle
	// chapters in vols
	total := 0
	for _, div2 := range ecfr.DIV1.DIV2 {
		if div2.TYPE != "SUBTITLE" {
			fmt.Println("Skipping", "not a subtitle", div2)
			continue
		}
		if div2.N != subtitle {
			continue
		}

		// chapters in subtitles
		for _, div3 := range div2.DIV3 {
			// ensure the type is "chapter"
			if div3.TYPE != "CHAPTER" {
				continue
			}

			// count the words in the chapter
			buf := new(bytes.Buffer)
			enc := xml.NewEncoder(buf)
			enc.Indent("", "")
			err := enc.Encode(div3)
			if err != nil {
				log.Fatal(err)
			}

			c, err := CountWords(bufio.NewReader(buf))
			if err != nil {
				return 0, err
			}

			total += c
		}
		return total, nil
	}

	return 0, fmt.Errorf("subtitle not found")
}

func (ecfr *ECFR) CountWordsPart(chapter string, part string) (int, error) {
	// find the chapter
	chap := (*DIV3)(nil)
	// chapters in vols
	for _, div2 := range ecfr.DIV1.DIV2 {
		if div2.TYPE != "SUBTITLE" {
			fmt.Println("Skipping", "not a subtitle", div2)
			continue
		}

		// chapters in subtitles
		for _, div3 := range div2.DIV3 {
			// get the chapter number
			chapterNum := div3.N
			if chapterNum != chapter {
				continue
			}
			// ensure the type is "chapter"
			if div3.TYPE != "CHAPTER" {
				continue
			}

			// count the words in the chapter
			buf := new(bytes.Buffer)
			enc := xml.NewEncoder(buf)
			enc.Indent("", "")
			err := enc.Encode(div3)
			if err != nil {
				log.Fatal(err)
			}

			chap = &div3

		}
	}

	if chap == nil {
		// chapters without vols
		for _, div3 := range ecfr.DIV1.DIV3 {
			// get the chapter number
			chapterNum := div3.N
			if chapterNum != chapter {
				continue
			}
			// ensure the type is "chapter"
			if div3.TYPE != "CHAPTER" {
				continue
			}

			// count the words in the chapter
			buf := new(bytes.Buffer)
			enc := xml.NewEncoder(buf)
			enc.Indent("", "")
			err := enc.Encode(div3)
			if err != nil {
				log.Fatal(err)
			}
			chap = &div3
		}
	}

	if chap == nil {
		fmt.Println("title", ecfr.DIV1.N, "chapter", chapter, "part", part, "not found")
		return 0, fmt.Errorf("chapter not found")
	}

	// find the part
	for _, div5 := range chap.DIV5 {
		if div5.TYPE != "PART" {
			continue
		}
		if div5.N != part {
			continue
		}
		// count the words in the part
		buf := new(bytes.Buffer)
		enc := xml.NewEncoder(buf)
		enc.Indent("", "")
		err := enc.Encode(div5)
		if err != nil {
			log.Fatal(err)
		}
		return CountWords(bufio.NewReader(buf))
	}

	// part not found
	// print out all debug details here
	fmt.Println("title", ecfr.DIV1.N, "chapter", chapter, "part", part, "not found")

	return 0, fmt.Errorf("part not found")
}

func (ecfr *ECFR) CountAllWords() (int, error) {
	// count all the words in the ecfr
	total := 0
	// chapters in vols
	for _, div2 := range ecfr.DIV1.DIV2 {
		if div2.TYPE != "SUBTITLE" {
			fmt.Println("Skipping", "not a subtitle", div2)
			continue
		}

		// chapters in subtitles
		for _, div3 := range div2.DIV3 {
			// ensure the type is "chapter"
			if div3.TYPE != "CHAPTER" {
				continue
			}

			// count the words in the chapter
			buf := new(bytes.Buffer)
			enc := xml.NewEncoder(buf)
			enc.Indent("", "")
			err := enc.Encode(div3)
			if err != nil {
				log.Fatal(err)
			}

			c, err := CountWords(bufio.NewReader(buf))
			if err != nil {
				return 0, err
			}

			total += c
		}
	}

	// chapters without vols
	for _, div3 := range ecfr.DIV1.DIV3 {
		// ensure the type is "chapter"
		if div3.TYPE != "CHAPTER" {
			continue
		}
		// count the words in the chapter
		buf := new(bytes.Buffer)
		enc := xml.NewEncoder(buf)
		enc.Indent("", "")
		err := enc.Encode(div3)
		if err != nil {
			log.Fatal(err)
		}
		c, err := CountWords(bufio.NewReader(buf))
		if err != nil {
			return 0, err
		}
		total += c
	}

	return total, nil
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
			DIV5 DIV5   `xml:"DIV5" json:"div5,omitempty"`

			DIV3 []DIV3 `xml:"DIV3" json:"div3,omitempty"`
		} `xml:"DIV2" json:"div2,omitempty"`
	} `xml:"DIV1" json:"div1,omitempty"`
}

type DIV3 struct {
	Text string `xml:",chardata" json:"text,omitempty"`
	N    string `xml:"N,attr" json:"n,omitempty"`
	TYPE string `xml:"TYPE,attr" json:"type,omitempty"`
	HEAD string `xml:"HEAD"`
	DIV5 []DIV5 `xml:"DIV5" json:"div5,omitempty"`
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

type DIV5 struct {
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
}

func NewECFR(filepath string) (*ECFR, error) {
	// open the file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	// read the file
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// unmarshal the XML data
	var ecfr ECFR
	err = xml.Unmarshal(data, &ecfr)
	if err != nil {
		return nil, err
	}

	return &ecfr, nil
}

func (ecfr *ECFR) Parse(data []byte) error {
	return xml.Unmarshal(data, ecfr)
}

// XMLWordCounter represents a decoder that counts words in XML content
type XMLWordCounter struct {
	WordCount int
}

// CountWords processes an XML file and returns the total word count
func CountWords(reader *bufio.Reader) (int, error) {

	counter := &XMLWordCounter{WordCount: 0}

	// Regular expressions for cleaning text
	tagRegex := regexp.MustCompile("<[^>]*>")
	spaceRegex := regexp.MustCompile(`\s+`)

	// Process the file line by line
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			// Log the error but continue processing
			fmt.Printf("Warning: error reading line: %v\n", err)
			continue
		}

		// Remove XML tags
		cleanText := tagRegex.ReplaceAllString(line, " ")

		// Remove XML entities
		cleanText = strings.ReplaceAll(cleanText, "&lt;", "<")
		cleanText = strings.ReplaceAll(cleanText, "&gt;", ">")
		cleanText = strings.ReplaceAll(cleanText, "&amp;", "&")
		cleanText = strings.ReplaceAll(cleanText, "&quot;", "\"")
		cleanText = strings.ReplaceAll(cleanText, "&apos;", "'")

		// Handle numeric entities
		numericEntityRegex := regexp.MustCompile(`&#\d+;`)
		cleanText = numericEntityRegex.ReplaceAllString(cleanText, " ")

		// Handle hex entities
		hexEntityRegex := regexp.MustCompile(`&#x[0-9a-fA-F]+;`)
		cleanText = hexEntityRegex.ReplaceAllString(cleanText, " ")

		// Normalize whitespace
		cleanText = spaceRegex.ReplaceAllString(cleanText, " ")
		cleanText = strings.TrimSpace(cleanText)

		// Count words in the cleaned text
		if cleanText != "" {
			words := strings.Fields(cleanText)
			counter.WordCount += len(words)
		}

		if err == io.EOF {
			break
		}
	}

	return counter.WordCount, nil
}
