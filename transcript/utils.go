package transcript

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// LoadCoursesFromJSONFile loads courses from a JSON file and converts them to Course structs
func LoadCoursesFromJSONFile(filename string) ([]Course, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var courses []Course
	err = json.Unmarshal(data, &courses)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return courses, nil
}

// ConvertRawJSONToCourses converts raw JSON data to Course structs
func ConvertRawJSONToCourses(data []byte) ([]Course, error) {
	var courses []Course
	err := json.Unmarshal(data, &courses)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return courses, nil
}

// GetCoursesBySemester filters courses by semester
func GetCoursesBySemester(courses []Course, semester string) []Course {
	var filtered []Course
	for _, course := range courses {
		if course.Semester == semester {
			filtered = append(filtered, course)
		}
	}
	return filtered
}

// GetCoursesByGrade filters courses by grade
func GetCoursesByGrade(courses []Course, grade string) []Course {
	var filtered []Course
	for _, course := range courses {
		if course.Grade == grade {
			filtered = append(filtered, course)
		}
	}
	return filtered
}

// CalculateGPASummary calculates GPA and credit summary from courses
func CalculateGPASummary(courses []Course) (float64, float64, int) {
	totalPoints := 0.0
	totalCredits := 0.0
	courseCount := 0

	gradePoints := map[string]float64{
		"AA": 4.0, "BA": 3.5, "BB": 3.0, "CB": 2.5,
		"CC": 2.0, "DC": 1.5, "DD": 1.0, "FD": 0.5,
		"FF": 0.0, "VF": 0.0, "BL": 0.0,
	}

	for _, course := range courses {
		credits, err := parseFloat(course.Credits)
		if err != nil {
			continue // Skip courses with invalid credits
		}

		points, exists := gradePoints[course.Grade]
		if !exists {
			continue // Skip courses with unknown grades
		}

		totalPoints += points * credits
		totalCredits += credits
		courseCount++
	}

	var gpa float64
	if totalCredits > 0 {
		gpa = totalPoints / totalCredits
	}

	return gpa, totalCredits, courseCount
}

// Helper function to parse string to float
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
} 