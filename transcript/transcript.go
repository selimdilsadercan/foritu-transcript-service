// Service transcript implements a PDF transcript parser REST API.
package transcript

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ledongthuc/pdf"
)

// TranscriptCourse represents a single course in the transcript
type TranscriptCourse struct {
	Semester string `json:"semester"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Credits  string `json:"credits"`
	Grade    string `json:"grade"`
}

// ParseTranscriptRequest represents the request body
type ParseTranscriptRequest struct {
	// PDF file content as base64 encoded string
	PDFBase64 string `json:"pdf_base64"`
}

// ParseTranscriptResponse represents the response
type ParseTranscriptResponse struct {
	Courses []TranscriptCourse `json:"courses"`
	Error   string             `json:"error,omitempty"`
	Debug   string             `json:"debug,omitempty"`
}

//encore:api public method=POST path=/parse-transcript
func ParseTranscript(ctx context.Context, req *ParseTranscriptRequest) (*ParseTranscriptResponse, error) {
	var debugInfo strings.Builder
	
	// Decode base64 PDF content
	pdfBytes, err := base64.StdEncoding.DecodeString(req.PDFBase64)
	if err != nil {
		return &ParseTranscriptResponse{
			Error: fmt.Sprintf("Failed to decode base64 PDF: %v", err),
		}, nil
	}

	debugInfo.WriteString(fmt.Sprintf("PDF decoded successfully, size: %d bytes\n", len(pdfBytes)))

	// Extract text from PDF
	text, err := extractTextFromPDF(pdfBytes)
	if err != nil {
		return &ParseTranscriptResponse{
			Error: fmt.Sprintf("Failed to extract text from PDF: %v", err),
		}, nil
	}

	debugInfo.WriteString(fmt.Sprintf("Text extracted successfully, length: %d characters\n", len(text)))
	
	// Debug: Add first 500 characters of extracted text to debug info
	previewLength := 500
	if len(text) < previewLength {
		previewLength = len(text)
	}
	debugInfo.WriteString(fmt.Sprintf("Text preview (first %d chars):\n%s\n", previewLength, text[:previewLength]))

	// Debug: Check if text was extracted
	if len(text) == 0 {
		return &ParseTranscriptResponse{
			Error: "No text extracted from PDF - PDF might be empty or unreadable",
		}, nil
	}

	// Parse the transcript text
	courses, parseDebug, err := parseTranscriptText(text)
	if err != nil {
		debugInfo.WriteString(fmt.Sprintf("Parse error: %v\n", err))
		debugInfo.WriteString(parseDebug)
		return &ParseTranscriptResponse{
			Error: fmt.Sprintf("Failed to parse transcript: %v", err),
			Debug: debugInfo.String(),
		}, nil
	}
	
	// Add parse debug info to main debug info
	debugInfo.WriteString(parseDebug)

	// Debug: Check if courses were found
	if len(courses) == 0 {
		return &ParseTranscriptResponse{
			Error: "No courses found in transcript. Check debug logs for details.",
			Debug: debugInfo.String(),
		}, nil
	}

	return &ParseTranscriptResponse{
		Courses: courses,
		Debug:   debugInfo.String(),
	}, nil
}

// extractTextFromPDF extracts text from PDF bytes
func extractTextFromPDF(pdfBytes []byte) (string, error) {
	// Create a reader for the PDF bytes
	reader := bytes.NewReader(pdfBytes)
	
	// Parse the PDF
	pdfReader, err := pdf.NewReader(reader, int64(len(pdfBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %v", err)
	}

	// Extract text from all pages
	var text bytes.Buffer
	for i := 1; i <= pdfReader.NumPage(); i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}
		
		pageText, err := page.GetPlainText(nil)
		if err != nil {
			continue // Skip pages that can't be read
		}
		

		
		text.WriteString(pageText)
		text.WriteString("\n")
	}

	return text.String(), nil
}

// parseTranscriptText parses the extracted text to find course information
func parseTranscriptText(text string) ([]TranscriptCourse, string, error) {
	var debugInfo strings.Builder
	debugInfo.WriteString(fmt.Sprintf("Starting to parse transcript text, length: %d\n", len(text)))
	
	// Search for semester patterns in the text - updated to include Yaz Okulu
	// Combine both patterns: regular semesters and Yaz Okulu
	semesterPattern := regexp.MustCompile(`(20\d{2}-20\d{2}\s+(Güz|Bahar|Yaz)\s+Dönemi|20\d{2}-20\d{2}\s+Yaz Okulu)`)
	debugInfo.WriteString(fmt.Sprintf("Searching for semester pattern: %s\n", semesterPattern.String()))
	semesterMatches := semesterPattern.FindAllStringIndex(text, -1)
	
	debugInfo.WriteString(fmt.Sprintf("Found %d semester matches\n", len(semesterMatches)))
	
	// Debug: Show what semester matches were found
	for i, match := range semesterMatches {
		semesterText := text[match[0]:match[1]]
		debugInfo.WriteString(fmt.Sprintf("Semester match %d: '%s'\n", i+1, semesterText))
	}
	
	// Debug: Show a sample of the text around the first semester match
	if len(semesterMatches) > 0 {
		firstMatch := semesterMatches[0]
		start := firstMatch[1]
		end := start + 200
		if end > len(text) {
			end = len(text)
		}
		sampleText := text[start:end]
		debugInfo.WriteString(fmt.Sprintf("Sample text after first semester: '%s'\n", sampleText))
	}
	
	// Debug: Show a larger sample of text to see the actual course format
	if len(semesterMatches) > 0 {
		// Look for a semester that might have courses
		for i, match := range semesterMatches {
			start := match[1]
			end := start + 500
			if end > len(text) {
				end = len(text)
			}
			semesterText := text[start:end]
			debugInfo.WriteString(fmt.Sprintf("Semester %d full text sample: '%s'\n", i+1, semesterText))
			// Only show first few semesters to avoid too much output
			if i >= 2 {
				break
			}
		}
	}
	
	// Debug: Check if semester patterns were found
	if len(semesterMatches) == 0 {
		debugInfo.WriteString("No semester matches found, trying alternative patterns\n")
		// Try alternative semester patterns that might be in the PDF
		altPatterns := []*regexp.Regexp{
			regexp.MustCompile(`(20\d{2}-20\d{2}\s+(Güz|Bahar|Yaz))`),
			regexp.MustCompile(`(20\d{2}\s+(Güz|Bahar|Yaz))`),
			regexp.MustCompile(`(Güz|Bahar|Yaz)\s+Dönemi`),
			regexp.MustCompile(`(Yaz Okulu)`),
		}
		
		for i, pattern := range altPatterns {
			debugInfo.WriteString(fmt.Sprintf("Trying alt pattern %d: %s\n", i+1, pattern.String()))
			matches := pattern.FindAllStringIndex(text, -1)
			debugInfo.WriteString(fmt.Sprintf("Alt pattern %d found %d matches\n", i+1, len(matches)))
			if len(matches) > 0 {
				// Found alternative pattern, use it
				semesterMatches = matches
				break
			}
		}
		
		// If still no matches, try to find any course codes and create a generic semester
		if len(semesterMatches) == 0 {
			debugInfo.WriteString("No semester patterns found, looking for course codes\n")
			courseCodePattern := regexp.MustCompile(`[A-Z]{3}\s+\d{3}[A-Z]*`)
			debugInfo.WriteString(fmt.Sprintf("Searching for course code pattern: %s\n", courseCodePattern.String()))
			courseMatches := courseCodePattern.FindAllStringIndex(text, -1)
			debugInfo.WriteString(fmt.Sprintf("Found %d course codes without semester\n", len(courseMatches)))
			if len(courseMatches) > 0 {
				// Found course codes but no semester, create a generic response
				return createGenericCourses(text), debugInfo.String(), nil
			}
			
			// No course codes found either
			return nil, debugInfo.String(), fmt.Errorf("no semester patterns or course codes found in text")
		}
	}
	
	var results []TranscriptCourse
	
	// Find all semester sections
	for i, semesterMatch := range semesterMatches {
		semester := text[semesterMatch[0]:semesterMatch[1]]
		startPos := semesterMatch[1]
		
		// Find the end of this semester's data (next semester or end of text)
		var endPos int
		if i+1 < len(semesterMatches) {
			endPos = semesterMatches[i+1][0]
		} else {
			endPos = len(text)
		}
		
		semesterText := text[startPos:endPos]
		
		// Clean up the semester text - remove header lines and summary lines
		lines := strings.Split(semesterText, "\n")
		var cleanedLines []string
		
		// Check if this is a Yaz Okulu semester - they have different formatting
		isYazOkulu := strings.Contains(semester, "Yaz Okulu")
		
		for _, line := range lines {
			// For Yaz Okulu semesters, be much more conservative with filtering
			if isYazOkulu {
				// For Yaz Okulu, only skip the most obvious header lines and keep everything else
				if strings.Contains(line, "Dersin Statüsü") || 
				   strings.Contains(line, "Öğretim Dili") || 
				   strings.Contains(line, "T U UK") || 
				   strings.Contains(line, "AKTS") || 
				   strings.Contains(line, "Not") || 
				   strings.Contains(line, "Puan") || 
				   strings.Contains(line, "Açıklama") ||
				   strings.Contains(line, "Öğrenci No") ||
				   strings.Contains(line, "T.C. Kimlik No") ||
				   strings.Contains(line, "Adı") ||
				   strings.Contains(line, "Doğum Tarihi") ||
				   strings.Contains(line, "Soyadı") ||
				   strings.Contains(line, "İSTANBUL TEKNİK ÜNİVERSİTESİ") ||
				   strings.Contains(line, "NOT DÖKÜM BELGESİ") ||
				   strings.Contains(line, "Belge Tarihi") ||
				   strings.Contains(line, "YOKTR") ||
				   strings.Contains(line, "www.turkiye.gov.tr") ||
				   strings.Contains(line, "Bu belgenin doğruluğunu") ||
				   strings.Contains(line, "SON SATIR") ||
				   strings.Contains(line, "Bu satırdan sonra") {
					continue
				}
			} else {
				// For regular semesters, use the original filtering logic
				if strings.Contains(line, "Dersin Statüsü") || 
				   strings.Contains(line, "Öğretim Dili") || 
				   strings.Contains(line, "T U UK") || 
				   strings.Contains(line, "AKTS") || 
				   strings.Contains(line, "Not") || 
				   strings.Contains(line, "Puan") || 
				   strings.Contains(line, "Açıklama") || 
				   strings.Contains(line, "DNO:") || 
				   strings.Contains(line, "GNO:") || 
				   strings.Contains(line, "TUK:") || 
				   strings.Contains(line, "TAKTS:") || 
				   strings.Contains(line, "DSD:") || 
				   strings.Contains(line, "Başarılı") || 
				   strings.Contains(line, "Pass") {
					continue
				}
			}
			
			if strings.TrimSpace(line) == "" {
				continue
			}
			cleanedLines = append(cleanedLines, line)
		}
		
		// If cleaned text is empty, try to preserve more content
		if len(cleanedLines) == 0 {
			// Try a less aggressive cleaning approach
			for _, line := range lines {
				// Only skip obvious header lines, keep everything else
				if strings.Contains(line, "Dersin Statüsü") || 
				   strings.Contains(line, "Öğretim Dili") || 
				   strings.Contains(line, "T U UK") || 
				   strings.Contains(line, "AKTS") || 
				   strings.Contains(line, "Not") || 
				   strings.Contains(line, "Puan") || 
				   strings.Contains(line, "Açıklama") {
					continue
				}
				if strings.TrimSpace(line) == "" {
					continue
				}
				cleanedLines = append(cleanedLines, line)
			}
		}
		
		// For Yaz Okulu semesters, if still no content, use raw text
		var cleanedText string
		if len(cleanedLines) == 0 && isYazOkulu {
			debugInfo.WriteString(fmt.Sprintf("DEBUG: Semester '%s' - No content after cleaning, using raw text\n", semester))
			// Use raw text for Yaz Okulu if cleaning removed everything
			cleanedText = semesterText
		} else {
			cleanedText = strings.Join(cleanedLines, "\n")
		}
		
		// Debug: Show the cleaned text for this semester
		debugInfo.WriteString(fmt.Sprintf("DEBUG: Semester '%s' - Cleaned text: '%s'\n", semester, cleanedText))
		
		// If cleaned text is empty, try to search in the raw semester text
		if len(cleanedText) == 0 {
			debugInfo.WriteString(fmt.Sprintf("DEBUG: Semester '%s' - Using raw text for course search\n", semester))
			cleanedText = semesterText
		}
		
		// Try to find course patterns in the cleaned text
		// Look for course codes with asterisk prefix and proper format
		// The pattern should match course codes like "ATA 121", "BLG 102E", "EKO 201E"
		// Use word boundaries to ensure we don't capture part of the course name
		courseCodePattern := regexp.MustCompile(`(\*?\s*[A-Z]{3}\s+\d{3}[A-Z]?)(?:\s|$)`)
		courseMatches := courseCodePattern.FindAllStringIndex(cleanedText, -1)
		
		// If no matches, try a simpler pattern
		if len(courseMatches) == 0 {
			courseCodePattern = regexp.MustCompile(`([A-Z]{3}\s+\d{3}[A-Z]?)(?:\s|$)`)
			courseMatches = courseCodePattern.FindAllStringIndex(cleanedText, -1)
		}
		
		// If still no matches, try a more flexible pattern that doesn't require word boundaries
		if len(courseMatches) == 0 {
			courseCodePattern = regexp.MustCompile(`([A-Z]{3}\s+\d{3}[A-Z]?)`)
			courseMatches = courseCodePattern.FindAllStringIndex(cleanedText, -1)
		}
		
		// For Yaz Okulu semesters, if still no matches, try searching in the raw text
		usingRawText := false
		if len(courseMatches) == 0 && isYazOkulu {
			debugInfo.WriteString(fmt.Sprintf("DEBUG: Semester '%s' - No course matches in cleaned text, trying raw text\n", semester))
			// Try to find course patterns in the raw semester text for Yaz Okulu
			courseCodePattern = regexp.MustCompile(`([A-Z]{3}\s+\d{3}[A-Z]?)`)
			courseMatches = courseCodePattern.FindAllStringIndex(semesterText, -1)
			debugInfo.WriteString(fmt.Sprintf("DEBUG: Semester '%s' - Found %d course matches in raw text\n", semester, len(courseMatches)))
			if len(courseMatches) > 0 {
				usingRawText = true
			}
			
			// If still no matches, try a more flexible pattern for Yaz Okulu
			if len(courseMatches) == 0 {
				debugInfo.WriteString(fmt.Sprintf("DEBUG: Semester '%s' - Trying more flexible pattern for Yaz Okulu\n", semester))
				// Try a more flexible pattern that might catch different formats
				flexiblePattern := regexp.MustCompile(`([A-Z]{2,4}\s+\d{2,4}[A-Z]?)`)
				courseMatches = flexiblePattern.FindAllStringIndex(semesterText, -1)
				debugInfo.WriteString(fmt.Sprintf("DEBUG: Semester '%s' - Found %d course matches with flexible pattern\n", semester, len(courseMatches)))
				if len(courseMatches) > 0 {
					usingRawText = true
				}
			}
		}
		
		debugInfo.WriteString(fmt.Sprintf("DEBUG: Semester '%s' - Found %d course matches with complex pattern\n", semester, len(courseMatches)))
		
		// Debug: Check if course patterns were found
		if len(courseMatches) == 0 {
			// Try a simpler course pattern
			simpleCoursePattern := regexp.MustCompile(`[A-Z]{3}\s+\d{3}[A-Z]*`)
			var textToSearch string
			if usingRawText {
				textToSearch = semesterText
			} else {
				textToSearch = cleanedText
			}
			simpleCourseMatches := simpleCoursePattern.FindAllStringIndex(textToSearch, -1)
			debugInfo.WriteString(fmt.Sprintf("DEBUG: Semester '%s' - Found %d course matches with simple pattern\n", semester, len(simpleCourseMatches)))
			if len(simpleCourseMatches) > 0 {
				// Use the simpler pattern results
				courseMatches = simpleCourseMatches	
			} else {
				debugInfo.WriteString(fmt.Sprintf("DEBUG: Semester '%s' - No course patterns found, skipping\n", semester))
				continue // Skip this semester if no course patterns found
			}
		}
		
		for j, courseMatch := range courseMatches {
			// Extract the course code, removing asterisk and extra spaces
			var sourceText string
			if usingRawText {
				sourceText = semesterText
			} else {
				sourceText = cleanedText
			}
			rawCode := strings.TrimSpace(strings.ReplaceAll(sourceText[courseMatch[0]:courseMatch[1]], "*", ""))
			
			// Debug: Show what was matched
			debugInfo.WriteString(fmt.Sprintf("DEBUG: Course match %d: rawCode='%s', match indices=[%d,%d]\n", j, rawCode, courseMatch[0], courseMatch[1]))
			
			// Clean up the course code - extract only the department code and course number
			// The course code should be in format like "ATA 121", "BLG 102E", "EKO 201E"
			var code string
			codeParts := strings.Fields(rawCode)
			if len(codeParts) >= 2 {
				// First part should be the department code (3 letters)
				// Second part should be the course number (3 digits + optional letter)
				if len(codeParts[0]) == 3 && len(codeParts[1]) >= 3 {
					// Extract just the course number part (3 digits + optional letter)
					courseNum := codeParts[1]
					// Find where the course number ends (3 digits + optional letter)
					courseNumPattern := regexp.MustCompile(`^\d{3}[A-Z]?`)
					if match := courseNumPattern.FindString(courseNum); match != "" {
						code = codeParts[0] + " " + match
					} else {
						// Fallback: keep only the department code and first 4 characters of course number
						if len(courseNum) >= 4 {
							code = codeParts[0] + " " + courseNum[:4]
						} else {
							code = codeParts[0] + " " + courseNum
						}
					}
				} else {
					code = rawCode
				}
			} else {
				code = rawCode
			}

			// Get the text after the course code
			startIdx := courseMatch[1]
			var endIdx int
			if j+1 < len(courseMatches) {
				endIdx = courseMatches[j+1][0]
			} else {
				if usingRawText {
					endIdx = len(semesterText)
				} else {
					endIdx = len(cleanedText)
				}
			}
			
			// For Yaz Okulu semesters, if we found matches in raw text, use raw text
			var courseText string
			if usingRawText {
				// Use raw semester text for extraction
				if j+1 < len(courseMatches) {
					endIdx = courseMatches[j+1][0]
				} else {
					endIdx = len(semesterText)
				}
				courseText = strings.TrimSpace(semesterText[startIdx:endIdx])
			} else {
				courseText = strings.TrimSpace(cleanedText[startIdx:endIdx])
			}
			
			debugInfo.WriteString(fmt.Sprintf("DEBUG: Processing course code '%s', course text length: %d\n", code, len(courseText)))
			
			// Clean up course text - remove footer content
			// Look for the first occurrence of footer indicators and cut off there
			footerIndicators := []string{
				"www.turkiye.gov.tr", "Öğrenci No", "T.C. Kimlik No", "SELİM DİLŞAD", 
				"ERCAN", "İSTANBUL TEKNİK ÜNİVERSİTESİ", "NOT DÖKÜM BELGESİ", 
				"YOKTR4QWO3AEMVO0BT", "Ders kodunun başında * olan dersler",
			}
			
			// For Yaz Okulu semesters, be more conservative with footer cleaning
			if !isYazOkulu {
				for _, indicator := range footerIndicators {
					if idx := strings.Index(courseText, indicator); idx != -1 {
						courseText = strings.TrimSpace(courseText[:idx])
						break
					}
				}
			} else {
				// For Yaz Okulu, only remove very obvious footer content
				strictFooterIndicators := []string{
					"www.turkiye.gov.tr", "NOT DÖKÜM BELGESİ", 
					"YOKTR4QWO3AEMVO0BT", "Ders kodunun başında * olan dersler",
				}
				for _, indicator := range strictFooterIndicators {
					if idx := strings.Index(courseText, indicator); idx != -1 {
						courseText = strings.TrimSpace(courseText[:idx])
						break
					}
				}
			}
			
					// Skip if no course data found - look for language patterns
		// Also check for garbled versions of the language patterns
		languagePattern := regexp.MustCompile(`(Tr|İng\.|Tr|İng)`)
		if !languagePattern.MatchString(courseText) {
			debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - No language pattern found, skipping\n", code))
			continue
		}
			
					// Try to extract course information using a more flexible approach
		// Look for the language pattern followed by numbers, allowing for newlines and flexible spacing
		// Pattern: Language + T U UK AKTS Grade Points Comment
		// Also handle garbled versions of the language patterns
		languageDataPattern := regexp.MustCompile(`(Tr|İng\.|Tr|İng)\s+(\d+)\s+(\d+)\s+(\d+\.?\d*)\s+(\d+\.?\d*)\s+(AA|BA\+?|BB\+?|CB\+?|CC\+?|DC\+?|DD\+?|BA|BB|CB|CC|DC|DD|FF|VF|BL|SG|DK|KL|--)\s+(\d+\.?\d*)(?:\s+([A-Z]{2}))?`)
		languageDataMatch := languageDataPattern.FindStringSubmatch(courseText)
		
		debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Language data match: %v\n", code, languageDataMatch != nil))
		
		// If the complex pattern fails, try a simpler approach
		if languageDataMatch == nil {
			// Try to find just the grade pattern
			gradePattern := regexp.MustCompile(`(AA|BA\+?|BB\+?|CB\+?|CC\+?|DC\+?|DD\+?|BA|BB|CB|CC|DC|DD|FF|VF|BL|SG|DK|KL|--)`)
			gradeMatch := gradePattern.FindString(courseText)
			if gradeMatch != "" {
				debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Found grade '%s' with simple pattern\n", code, gradeMatch))
				
				// --- UK COLUMN CREDIT EXTRACTION ---
				// Look for the UK column value in the table structure
				// Pattern: language followed by numbers (T, U, UK, AKTS columns)
				// Format: İng.32488 or Tr20020 (language + numbers stuck together)
				ukCreditPattern := regexp.MustCompile(`(Tr|İng\.)(\d{1})(\d{1})(\d{1})(\d{1,2})`)
				ukCreditMatch := ukCreditPattern.FindStringSubmatch(courseText)
				var credits string
				if ukCreditMatch != nil && len(ukCreditMatch) >= 6 {
					// Extract the UK column value (4th capture group)
					ukValue := ukCreditMatch[4]
					if len(ukValue) > 0 {
						// Check if this is a Turkish course that should have 0 credits
						// ATA and TUR courses typically have 0 credits
						if strings.HasPrefix(code, "ATA ") || strings.HasPrefix(code, "TUR ") {
							credits = "0"
							debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Turkish course (ATA/TUR) detected, setting credits to '0'\n", code))
						} else {
							// For all other courses, use the UK value as credits
							// Clean up the number to ensure it's valid
							cleanNumber := regexp.MustCompile(`[^0-9.]`).ReplaceAllString(ukValue, "")
							if cleanNumber == "" || cleanNumber == "." {
								credits = "0"
							} else {
								// Validate that it's a reasonable credit value (0-10 range)
								if f, err := strconv.ParseFloat(cleanNumber, 64); err == nil && f >= 0 && f <= 10 {
									credits = cleanNumber
								} else {
									// If the UK value is not valid, try to extract a reasonable credit value
									// Look for common credit patterns (1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
									creditValuePattern := regexp.MustCompile(`([0-9]|10)`)
									if match := creditValuePattern.FindString(cleanNumber); match != "" {
										credits = match
									} else {
										credits = "0"
									}
								}
							}
						}
					} else {
						credits = "0"
					}
					debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Found UK value: '%s', extracted credits: '%s'\n", code, ukValue, credits))
				} else {
					// Fallback to old pattern if UK pattern doesn't match
					creditPattern := regexp.MustCompile(`(Tr|İng\.)\s*([0-9]+\.?[0-9]*)`)
					creditMatch := creditPattern.FindStringSubmatch(courseText)
					if creditMatch != nil && len(creditMatch) >= 3 {
						fullNumber := creditMatch[2]
						if len(fullNumber) > 0 {
							if strings.HasPrefix(code, "ATA ") || strings.HasPrefix(code, "TUR ") {
								credits = "0"
							} else {
								cleanNumber := regexp.MustCompile(`[^0-9.]`).ReplaceAllString(fullNumber, "")
								if cleanNumber == "" || cleanNumber == "." {
									credits = "0"
								} else {
									if f, err := strconv.ParseFloat(cleanNumber, 64); err == nil && f >= 0 && f <= 10 {
										credits = cleanNumber
									} else {
										creditValuePattern := regexp.MustCompile(`([0-9]|10)`)
										if match := creditValuePattern.FindString(cleanNumber); match != "" {
											credits = match
										} else {
											credits = "0"
										}
									}
								}
							}
						} else {
							credits = "0"
						}
						debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Fallback: Found full number: '%s', extracted credits: '%s'\n", code, fullNumber, credits))
					} else {
						credits = "0"
						debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - No credits found with any pattern, defaulting to '0'\n", code))
					}
				}
				
				// Try to extract course name (everything before the language code)
				name := "Unknown Course"
				
				// First, try to find the language pattern and extract everything before it
				langPattern := regexp.MustCompile(`(Tr|İng\.)`)
				langMatches := langPattern.FindAllStringIndex(courseText, -1)
				debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Language matches: %v\n", code, langMatches))
				debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Course text: '%s'\n", code, courseText))
				if len(langMatches) > 0 {
					// Use the first language occurrence (usually the one after the course name)
					langMatch := langMatches[0]
					namePart := courseText[:langMatch[0]]
					debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Name part before language: '%s'\n", code, namePart))
					name = strings.TrimSpace(namePart)
					
					// Clean up the name - remove English translations in parentheses and newlines
					name = regexp.MustCompile(`\s*\([^)]*\)\s*`).ReplaceAllString(name, "")
					name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ") // Replace multiple spaces/newlines with single space
					name = strings.TrimSpace(name)
					
					// Remove trailing parentheses that shouldn't be there
					name = strings.TrimSuffix(name, ")")
					name = strings.TrimSuffix(name, "(")
					debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Cleaned name: '%s'\n", code, name))
					
					// If name is empty, try to extract from parentheses
					if name == "" {
						parenPattern := regexp.MustCompile(`\(([^)]+)\)`)
						parenMatches := parenPattern.FindStringSubmatch(courseText)
						if len(parenMatches) > 1 {
							name = strings.TrimSpace(parenMatches[1])
						}
					}
				} else {
					// Fallback: try to extract from before the grade
					if gradeMatch != "" {
						namePart := courseText[:strings.Index(courseText, gradeMatch)]
						name = strings.TrimSpace(namePart)
						if name == "" {
							name = "Unknown Course"
						}
					}
				}
				
				// Remove 'L' prefix from laboratory course names
				// Laboratory courses have 'L' at the beginning of the name
				if strings.HasPrefix(name, "L") && len(name) > 1 {
					// Check if the second character is uppercase (likely part of the course name)
					if len(name) > 1 && name[1] >= 'A' && name[1] <= 'Z' {
						name = name[1:] // Remove the 'L' prefix
						debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Laboratory course detected, removed 'L' prefix from name: '%s'\n", code, name))
					}
				}
				
				// Check if this is a Turkish course or specific English course (ING 100E) and correct the course code
				finalCode := code
				finalName := name
				if strings.Contains(courseText, "Tr") || (strings.HasPrefix(code, "ING 100") && strings.Contains(courseText, "İng.")) {
					// Extract department code and course number without letter suffix
					codeParts := strings.Fields(code)
					if len(codeParts) >= 2 {
						deptCode := codeParts[0]
						courseNum := codeParts[1]
						// Remove letter suffix from course number for Turkish courses and ING 100E
						courseNumPattern := regexp.MustCompile(`^\d{3}`)
						if match := courseNumPattern.FindString(courseNum); match != "" {
							finalCode = deptCode + " " + match
							// Add the removed letter to the beginning of the course name
							if len(courseNum) > 3 {
								removedLetter := courseNum[3:4] // Get the letter after the 3 digits
								finalName = removedLetter + name
								debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - %s course detected, corrected code to '%s', name to '%s'\n", code, func() string { if strings.Contains(courseText, "Tr") { return "Turkish" } else { return "ING 100" } }(), finalCode, finalName))
							}
						}
					}
				}
				
				// Check if this is a laboratory course and add 'L' suffix to course code
				if strings.Contains(strings.ToLower(name), "laboratory") || strings.Contains(strings.ToLower(name), "lab") {
					// Add 'L' suffix to the course code if it doesn't already have it
					if !strings.HasSuffix(finalCode, "L") {
						finalCode = finalCode + "L"
						debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Laboratory course detected, added 'L' suffix to code: '%s'\n", code, finalCode))
					}
				}
				
				results = append(results, TranscriptCourse{
					Semester: semester,
					Code:     finalCode,
					Name:     finalName,
					Credits:  credits,
					Grade:    gradeMatch,
				})
				continue
			}
		}
		
		if languageDataMatch != nil {
				language := strings.TrimSpace(languageDataMatch[1])
				localCredits := languageDataMatch[4]
				grade := strings.TrimSpace(languageDataMatch[6])
				
				// Debug: Log all captured groups
				debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Language data match groups: %v\n", code, languageDataMatch))
				debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Raw localCredits: '%s'\n", code, localCredits))
				
				// Everything before the language is the course name
				namePart := courseText[:strings.Index(courseText, language)]
				namePart = strings.TrimSpace(namePart)
				
				// Clean up the name - remove English translations in parentheses and newlines
				name := regexp.MustCompile(`\s*\([^)]*\)\s*`).ReplaceAllString(namePart, "")
				name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ") // Replace multiple spaces/newlines with single space
				name = strings.TrimSpace(name)
				
				// If name is empty, try to extract from parentheses
				if name == "" {
					parenPattern := regexp.MustCompile(`\(([^)]+)\)`)
					parenMatches := parenPattern.FindStringSubmatch(courseText)
					if len(parenMatches) > 1 {
						name = strings.TrimSpace(parenMatches[1])
					}
				}
				
				// Remove 'L' prefix from laboratory course names
				// Laboratory courses have 'L' at the beginning of the name
				if strings.HasPrefix(name, "L") && len(name) > 1 {
					// Check if the second character is uppercase (likely part of the course name)
					if len(name) > 1 && name[1] >= 'A' && name[1] <= 'Z' {
						name = name[1:] // Remove the 'L' prefix
					}
				}
				
				// Use the local credits (usually the smaller number)
				// Clean up the credits - should be a simple number like 0, 1, 2, 3, 4
				credits := strings.TrimSpace(localCredits)
				// Remove any non-digit characters except decimal point
				credits = regexp.MustCompile(`[^0-9.]`).ReplaceAllString(credits, "")
				// If credits is empty or invalid, default to "0"
				if credits == "" || credits == "." {
					credits = "0"
				}
				
				// Debug: Log what we're extracting
				debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Cleaned credits: '%s'\n", code, credits))
				
				// Check if this is a Turkish course and correct the course code
				finalCode := code
				finalName := name
				if strings.Contains(courseText, "Tr") {
					// Extract department code and course number without letter suffix
					codeParts := strings.Fields(code)
					if len(codeParts) >= 2 {
						deptCode := codeParts[0]
						courseNum := codeParts[1]
						// Remove letter suffix from course number for Turkish courses
						courseNumPattern := regexp.MustCompile(`^\d{3}`)
						if match := courseNumPattern.FindString(courseNum); match != "" {
							finalCode = deptCode + " " + match
							// Add the removed letter to the beginning of the course name
							if len(courseNum) > 3 {
								removedLetter := courseNum[3:4] // Get the letter after the 3 digits
								finalName = removedLetter + name
							}
						}
					}
				}
				
				// Check if this is a laboratory course and add 'L' suffix to course code
				if strings.Contains(strings.ToLower(name), "laboratory") || strings.Contains(strings.ToLower(name), "lab") {
					// Add 'L' suffix to the course code if it doesn't already have it
					if !strings.HasSuffix(finalCode, "L") {
						finalCode = finalCode + "L"
					}
				}
				
				results = append(results, TranscriptCourse{
					Semester: semester,
					Code:     finalCode,
					Name:     finalName,
					Credits:  credits,
					Grade:    grade,
				})
			} else {
				// Try a simpler approach - just find the language and then look for numbers
				// Find all language occurrences
				langPattern := regexp.MustCompile(`(Tr|İng\.)`)
				langMatches := langPattern.FindAllStringIndex(courseText, -1)
				debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Language pattern matches: %v\n", code, langMatches))
				if len(langMatches) > 0 {
					// Use the last language occurrence (usually the one before the data)
					langMatch := langMatches[len(langMatches)-1]
					language := courseText[langMatch[0]:langMatch[1]]
					language = strings.TrimSpace(language)
					
					// Get text after language
					afterLang := strings.TrimSpace(courseText[langMatch[1]:])
					
					// Try to extract numbers manually by splitting
					parts := strings.Fields(afterLang)
					debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Parts after language: %v\n", code, parts))
					if len(parts) >= 6 {
						localCredits := parts[2]
						debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Raw localCredits from parts: '%s'\n", code, localCredits))
						
						// Check if parts[4] is a grade (letter) or points (number)
						grade := parts[4]
						points := parts[5]
						
						// Validate grade format
						validGrades := []string{"AA", "BA+", "BA", "BB+", "BB", "CB+", "CB", "CC+", "CC", "DC+", "DC", "DD+", "DD", "FF", "VF", "BL", "SG", "DK", "KL", "--"}
						isValidGrade := false
						for _, validGrade := range validGrades {
							if grade == validGrade {
								isValidGrade = true
								break
							}
						}
						
						if !isValidGrade {
							// They might be swapped
							points, grade = grade, points
						}
						
						// Everything before the language is the course name
						namePart := strings.TrimSpace(courseText[:langMatch[0]])
						name := regexp.MustCompile(`\s*\([^)]*\)\s*`).ReplaceAllString(namePart, "")
						name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
						name = strings.TrimSpace(name)
						
						// If name is empty, try to extract from parentheses
						if name == "" {
							parenPattern := regexp.MustCompile(`\(([^)]+)\)`)
							parenMatches := parenPattern.FindStringSubmatch(courseText)
							if len(parenMatches) > 1 {
								name = strings.TrimSpace(parenMatches[1])
							}
						}
						
						// Remove 'L' prefix from laboratory course names
						// Laboratory courses have 'L' at the beginning of the name
						if strings.HasPrefix(name, "L") && len(name) > 1 {
							// Check if the second character is uppercase (likely part of the course name)
							if len(name) > 1 && name[1] >= 'A' && name[1] <= 'Z' {
								name = name[1:] // Remove the 'L' prefix
								debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Laboratory course detected, removed 'L' prefix from name: '%s'\n", code, name))
							}
						}
						
						// Extract credits from the parts - look for the smallest number that could be credits
						// Credits are usually 0, 1, 2, 3, 4, or decimal values like 1.5, 3.5
						var credits string
						debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Looking for credits in parts: %v\n", code, parts))
						for i, part := range parts {
							// Clean the part to get just numbers and decimal points
							cleanPart := regexp.MustCompile(`[^0-9.]`).ReplaceAllString(part, "")
							debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Part %d: '%s' -> clean: '%s'\n", code, i, part, cleanPart))
							if cleanPart != "" && cleanPart != "." {
								// Check if this looks like a credit value (0-10 range, possibly decimal)
								if f, err := strconv.ParseFloat(cleanPart, 64); err == nil && f >= 0 && f <= 10 {
									credits = cleanPart
									debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Found valid credits: '%s' (float: %f)\n", code, credits, f))
									break
								} else {
									debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Part %d not valid credits: err=%v, f=%f\n", code, i, err, f))
								}
							}
						}
						
						// If no valid credits found, default to "0"
						if credits == "" {
							credits = "0"
							debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - No valid credits found, defaulting to '0'\n", code))
						}
						
						debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Final credits: '%s'\n", code, credits))
						debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - About to check for Turkish course\n", code))
						
						// Check if this is a Turkish course and correct the course code
						// For Turkish courses (marked with "Tr"), remove the letter suffix from course code
						// The letter is part of the course name, not the course code
						finalCode := code
						finalName := name
						
						// Check if this is a Turkish course by looking at the language matches
						// Turkish courses have "Tr" in their language pattern
						languagePattern := regexp.MustCompile(`(Tr|İng\.|Tr|İng)`)
						langMatches := languagePattern.FindAllStringIndex(courseText, -1)
						isTurkishCourse := false
						
						debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Checking for Turkish course, found %d language matches\n", code, len(langMatches)))
						
						for _, langMatch := range langMatches {
							langText := courseText[langMatch[0]:langMatch[1]]
							debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Language match: '%s'\n", code, langText))
							if langText == "Tr" {
								isTurkishCourse = true
								debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Found Turkish language marker\n", code))
								break
							}
						}
						
						if isTurkishCourse {
							// Extract department code and course number without letter suffix
							codeParts := strings.Fields(code)
							if len(codeParts) >= 2 {
								deptCode := codeParts[0]
								courseNum := codeParts[1]
								// Remove letter suffix from course number for Turkish courses
								courseNumPattern := regexp.MustCompile(`^\d{3}`)
								if match := courseNumPattern.FindString(courseNum); match != "" {
									finalCode = deptCode + " " + match
									// Add the removed letter to the beginning of the course name
									if len(courseNum) > 3 {
										removedLetter := courseNum[3:4] // Get the letter after the 3 digits
										finalName = removedLetter + name
										debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Turkish course detected, corrected code to '%s', name to '%s'\n", code, finalCode, finalName))
									} else {
										debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Turkish course detected, corrected code to '%s'\n", code, finalCode))
									}
								}
							}
						}
						
						// Check if this is a laboratory course and add 'L' suffix to course code
						if strings.Contains(strings.ToLower(name), "laboratory") || strings.Contains(strings.ToLower(name), "lab") {
							// Add 'L' suffix to the course code if it doesn't already have it
							if !strings.HasSuffix(finalCode, "L") {
								finalCode = finalCode + "L"
								debugInfo.WriteString(fmt.Sprintf("DEBUG: Course '%s' - Laboratory course detected, added 'L' suffix to code: '%s'\n", code, finalCode))
							}
						}
						
						results = append(results, TranscriptCourse{
							Semester: semester,
							Code:     finalCode,
							Name:     name,
							Credits:  credits,
							Grade:    grade,
						})
					}
				}
			}
		}
	}
	
	debugInfo.WriteString(fmt.Sprintf("Total courses found: %d\n", len(results)))
	return results, debugInfo.String(), nil
}

// createGenericCourses creates courses when semester information is not found
func createGenericCourses(text string) []TranscriptCourse {
	var results []TranscriptCourse
	
	// Find all course codes in the text
	courseCodePattern := regexp.MustCompile(`(\*?\s*[A-Z]{3}\s+\d{3}[A-Z]*)`)
	courseMatches := courseCodePattern.FindAllStringIndex(text, -1)
	
	for i, courseMatch := range courseMatches {
		code := text[courseMatch[0]:courseMatch[1]]
		
		// Get text after the course code
		startIdx := courseMatch[1]
		var endIdx int
		if i+1 < len(courseMatches) {
			endIdx = courseMatches[i+1][0]
		} else {
			endIdx = len(text)
		}
		
		courseText := strings.TrimSpace(text[startIdx:endIdx])
		
		// Try to extract basic course information
		// Look for common patterns in the course text
		gradePattern := regexp.MustCompile(`(AA|BA\+?|BB\+?|CB\+?|CC\+?|DC\+?|DD\+?|BA|BB|CB|CC|DC|DD|FF|VF|BL|SG|DK|KL|--)`)
		gradeMatch := gradePattern.FindString(courseText)
		
		// Look for credit patterns (numbers that could be credits)
		creditPattern := regexp.MustCompile(`(\d+\.?\d*)`)
		creditMatches := creditPattern.FindAllString(courseText, -1)
		
		grade := "N/A"
		if gradeMatch != "" {
			grade = gradeMatch
		}
		
		credits := "N/A"
		if len(creditMatches) > 0 {
			credits = creditMatches[0] // Use first number as credits
		}
		
		// Try to extract course name (everything before the first grade or credit)
		name := "Unknown Course"
		if gradeMatch != "" {
			namePart := courseText[:strings.Index(courseText, gradeMatch)]
			name = strings.TrimSpace(namePart)
			if name == "" {
				name = "Unknown Course"
			}
		}
		
		// Remove trailing parentheses that shouldn't be there
		name = strings.TrimSuffix(name, ")")
		name = strings.TrimSuffix(name, "(")
		
		// Remove 'L' prefix from laboratory course names
		// Laboratory courses have 'L' at the beginning of the name
		if strings.HasPrefix(name, "L") && len(name) > 1 {
			// Check if the second character is uppercase (likely part of the course name)
			if len(name) > 1 && name[1] >= 'A' && name[1] <= 'Z' {
				name = name[1:] // Remove the 'L' prefix
			}
		}
		
						// Check if this is a Turkish or English course and correct the course code
				// For Turkish courses (marked with "Tr") and English courses (marked with "İng."), remove the letter suffix from course code
				finalCode := code
				finalName := name
				if strings.Contains(courseText, "Tr") || strings.Contains(courseText, "İng.") {
					// Extract department code and course number without letter suffix
					codeParts := strings.Fields(code)
					if len(codeParts) >= 2 {
						deptCode := codeParts[0]
						courseNum := codeParts[1]
						// Remove letter suffix from course number for Turkish and English courses
						courseNumPattern := regexp.MustCompile(`^\d{3}`)
						if match := courseNumPattern.FindString(courseNum); match != "" {
							finalCode = deptCode + " " + match
							// Add the removed letter to the beginning of the course name
							if len(courseNum) > 3 {
								removedLetter := courseNum[3:4] // Get the letter after the 3 digits
								finalName = removedLetter + name
							}
						}
					}
				}
		
		// Check if this is a laboratory course and add 'L' suffix to course code
		if strings.Contains(strings.ToLower(name), "laboratory") || strings.Contains(strings.ToLower(name), "lab") {
			// Add 'L' suffix to the course code if it doesn't already have it
			if !strings.HasSuffix(finalCode, "L") {
				finalCode = finalCode + "L"
			}
		}
		
		results = append(results, TranscriptCourse{
			Semester: "Unknown Semester",
			Code:     finalCode,
			Name:     finalName,
			Credits:  credits,
			Grade:    grade,
		})
	}
	
	return results
}

 