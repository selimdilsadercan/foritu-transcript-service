import re
import json
from PyPDF2 import PdfReader

def extract_text_from_pdf(file_path):
    reader = PdfReader(file_path)
    text = ""
    for page in reader.pages:
        text += page.extract_text()
    return text

def parse_transcript(text):
    # Search for semester patterns in the text - updated to include Yaz Okulu
    # Combine both patterns: regular semesters and Yaz Okulu
    semester_pattern = r"(20\d{2}-20\d{2}\s+(Güz|Bahar|Yaz)\s+Dönemi|20\d{2}-20\d{2}\s+Yaz Okulu)"
    semester_matches = list(re.finditer(semester_pattern, text))
    
    results = []
    
    # Find all semester sections
    for i, semester_match in enumerate(semester_matches):
        semester = semester_match.group(0)
        start_pos = semester_match.end()
        
        # Find the end of this semester's data (next semester or end of text)
        if i + 1 < len(semester_matches):
            end_pos = semester_matches[i + 1].start()
        else:
            end_pos = len(text)
        
        semester_text = text[start_pos:end_pos]
        
        # Clean up the semester text - remove header lines and summary lines
        lines = semester_text.split('\n')
        cleaned_lines = []
        for line in lines:
            # Skip header lines and summary lines
            if any(skip in line for skip in ['Dersin Statüsü', 'Öğretim Dili', 'T U UK', 'AKTS', 'Not', 'Puan', 'Açıklama', 'DNO:', 'GNO:', 'TUK:', 'TAKTS:', 'DSD:', 'Başarılı', 'Pass']):
                continue
            if line.strip() == '':
                continue
            cleaned_lines.append(line)
        
        cleaned_text = '\n'.join(cleaned_lines)
        
        # Try to find course patterns in the cleaned text
        # Updated pattern to properly capture course codes with letters at the end like FIZ 102EL
        course_code_pattern = r"(\*?\s*[A-Z]{3}\s+\d{3}[A-Z]*)\s"
        course_matches = list(re.finditer(course_code_pattern, cleaned_text))
        
        for j, course_match in enumerate(course_matches):
            code = course_match.group(1).replace("*", "").strip()
            
            # Get the text after the course code
            start_idx = course_match.end()
            if j + 1 < len(course_matches):
                end_idx = course_matches[j + 1].start()
            else:
                end_idx = len(cleaned_text)
            
            course_text = cleaned_text[start_idx:end_idx].strip()
            
            # Clean up course text - remove footer content
            # Look for the first occurrence of footer indicators and cut off there
            footer_indicators = ['www.turkiye.gov.tr', 'Öğrenci No', 'T.C. Kimlik No', 'SELİM DİLŞAD', 'ERCAN', 'İSTANBUL TEKNİK ÜNİVERSİTESİ', 'NOT DÖKÜM BELGESİ', 'YOKTR4QWO3AEMVO0BT', 'Ders kodunun başında * olan dersler']
            
            for indicator in footer_indicators:
                if indicator in course_text:
                    course_text = course_text[:course_text.find(indicator)].strip()
                    break
            
            # Skip if no course data found
            if not re.search(r'(Tr|İng\.)\s+\d+\s+\d+\s+\d+', course_text):
                continue
            
            # Try to extract course information using a more flexible approach
            # The course text contains: course name, language, and numerical data
            
            # Look for the language pattern followed by numbers, allowing for newlines and flexible spacing
            # Pattern: Language + T U UK AKTS Grade Points Comment
            language_data_pattern = r"(Tr|İng\.)\s+(\d+)\s+(\d+)\s+(\d+\.?\d*)\s+(\d+\.?\d*)\s+(AA|BA\+?|BB\+?|CB\+?|CC\+?|DC\+?|DD\+?|BA|BB|CB|CC|DC|DD|FF|VF|BL|SG|DK|KL|--)\s+(\d+\.?\d*)(?:\s+([A-Z]{2}))?"
            language_data_match = re.search(language_data_pattern, course_text, re.DOTALL)
            
            if language_data_match:
                language = language_data_match.group(1).strip()
                theory_hours = language_data_match.group(2)
                lab_hours = language_data_match.group(3)
                local_credits = language_data_match.group(4)
                ects_credits = language_data_match.group(5)
                grade = language_data_match.group(6).strip()
                points = language_data_match.group(7)
                comment = language_data_match.group(8) if language_data_match.group(8) else ""
                
                # Everything before the language is the course name
                name_part = course_text[:language_data_match.start()].strip()
                
                # Clean up the name - remove English translations in parentheses and newlines
                name = re.sub(r'\s*\([^)]*\)\s*', '', name_part).strip()
                name = re.sub(r'\s+', ' ', name)  # Replace multiple spaces/newlines with single space
                
                results.append({
                    "semester": semester,
                    "code": code,
                    "name": name,
                    "language": language,
                    "theory_hours": theory_hours,
                    "lab_hours": lab_hours,
                    "local_credits": local_credits,
                    "ects_credits": ects_credits,
                    "grade": grade,
                    "points": points,
                    "comment": comment
                })
            else:
                # Try a simpler approach - just find the language and then look for numbers
                # Find all language occurrences
                lang_matches = list(re.finditer(r"(Tr|İng\.)", course_text))
                if lang_matches:
                    # Use the last language occurrence (usually the one before the data)
                    lang_match = lang_matches[-1]
                    language = lang_match.group(1).strip()
                    
                    # Get text after language
                    after_lang = course_text[lang_match.end():].strip()
                    
                    # Try to extract numbers manually by splitting
                    parts = after_lang.split()
                    if len(parts) >= 6:
                        try:
                            theory_hours = parts[0]
                            lab_hours = parts[1]
                            local_credits = parts[2]
                            ects_credits = parts[3]
                            
                            # Check if parts[4] is a grade (letter) or points (number)
                            if parts[4] in ['AA', 'BA+', 'BA', 'BB+', 'BB', 'CB+', 'CB', 'CC+', 'CC', 'DC+', 'DC', 'DD+', 'DD', 'FF', 'VF', 'BL', 'SG', 'DK', 'KL', '--']:
                                grade = parts[4]
                                points = parts[5]
                            else:
                                # They might be swapped
                                points = parts[4]
                                grade = parts[5]
                            
                            comment = parts[6] if len(parts) > 6 else ""
                            
                            # Everything before the language is the course name
                            name_part = course_text[:lang_match.start()].strip()
                            name = re.sub(r'\s*\([^)]*\)\s*', '', name_part).strip()
                            name = re.sub(r'\s+', ' ', name)
                            
                            results.append({
                                "semester": semester,
                                "code": code,
                                "name": name,
                                "language": language,
                                "theory_hours": theory_hours,
                                "lab_hours": lab_hours,
                                "local_credits": local_credits,
                                "ects_credits": ects_credits,
                                "grade": grade,
                                "points": points,
                                "comment": comment
                            })
                        except Exception as e:
                            pass  # Skip this course if there's an error
    
    return results

def create_simple_output(courses):
    """Create a simple output with only essential fields"""
    simple_courses = []
    for course in courses:
        simple_courses.append({
            "semester": course["semester"],
            "code": course["code"],
            "name": course["name"],
            "credits": course["local_credits"],
            "grade": course["grade"]
        })
    return simple_courses

if __name__ == "__main__":
    pdf_path = "transkript.pdf"  # Dosyanın adı aynı dizinde olmalı
    text = extract_text_from_pdf(pdf_path)
    courses = parse_transcript(text)
    
    # Create simple output
    simple_courses = create_simple_output(courses)
    
    print(f"Found {len(courses)} courses:")
    print("\n=== SIMPLE OUTPUT ===")
    print(json.dumps(simple_courses, indent=2, ensure_ascii=False))
    
    # Save simple version to JSON file
    with open('transcript_simple.json', 'w', encoding='utf-8') as f:
        json.dump(simple_courses, f, indent=2, ensure_ascii=False)
    print(f"\nSimple courses saved to transcript_simple.json")
