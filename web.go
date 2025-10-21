package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

// Response represents the JSON response for API calls
type Response struct {
	Success bool     `json:"success"`
	Message string   `json:"message,omitempty"`
	INNs    []string `json:"inns,omitempty"`
	Valid   bool     `json:"valid,omitempty"`
}

var htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>INN Generator and Validator</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 50px auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        h1 {
            color: #333;
            text-align: center;
        }
        .section {
            background: white;
            padding: 20px;
            margin: 20px 0;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h2 {
            color: #555;
            border-bottom: 2px solid #007bff;
            padding-bottom: 10px;
        }
        .form-group {
            margin: 15px 0;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
        }
        input[type="text"], input[type="number"] {
            width: 100%;
            padding: 8px;
            border: 1px solid #ddd;
            border-radius: 4px;
            box-sizing: border-box;
        }
        button {
            background-color: #007bff;
            color: white;
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
        }
        button:hover {
            background-color: #0056b3;
        }
        .result {
            margin-top: 15px;
            padding: 10px;
            border-radius: 4px;
        }
        .result.success {
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .result.error {
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
        .inn-list {
            list-style: none;
            padding: 0;
        }
        .inn-list li {
            padding: 8px;
            margin: 5px 0;
            background-color: #f8f9fa;
            border-radius: 4px;
            border-left: 3px solid #007bff;
        }
    </style>
</head>
<body>
    <h1>INN Generator and Validator</h1>
    <p style="text-align: center; color: #666;">Taxpayer Identification Number (INN) tool</p>

    <div class="section">
        <h2>Validate INN</h2>
        <div class="form-group">
            <label for="checkINN">Enter INN to validate (10 or 12 digits):</label>
            <input type="text" id="checkINN" placeholder="e.g., 7707083893">
            <button onclick="validateINN()">Validate</button>
        </div>
        <div id="validateResult"></div>
    </div>

    <div class="section">
        <h2>Generate INN for Physical Person</h2>
        <div class="form-group">
            <label for="physicalCount">Number of INNs to generate:</label>
            <input type="number" id="physicalCount" value="5" min="1" max="100">
            <button onclick="generatePhysical()">Generate</button>
        </div>
        <div id="physicalResult"></div>
    </div>

    <div class="section">
        <h2>Generate INN for Juridical Person</h2>
        <div class="form-group">
            <label for="juridicalCount">Number of INNs to generate:</label>
            <input type="number" id="juridicalCount" value="5" min="1" max="100">
            <button onclick="generateJuridical()">Generate</button>
        </div>
        <div id="juridicalResult"></div>
    </div>

    <script>
        function validateINN() {
            const inn = document.getElementById('checkINN').value;
            const resultDiv = document.getElementById('validateResult');
            
            if (!inn) {
                resultDiv.innerHTML = '<div class="result error">Please enter an INN</div>';
                return;
            }

            fetch('/api/validate?inn=' + encodeURIComponent(inn))
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        const innType = inn.length === 12 ? 'physical' : 'juridical';
                        const resultClass = data.valid ? 'success' : 'error';
                        const resultText = data.valid 
                            ? 'INN ' + inn + ' is valid (' + innType + ' person)'
                            : 'INN ' + inn + ' is invalid';
                        resultDiv.innerHTML = '<div class="result ' + resultClass + '">' + resultText + '</div>';
                    } else {
                        resultDiv.innerHTML = '<div class="result error">' + data.message + '</div>';
                    }
                })
                .catch(error => {
                    resultDiv.innerHTML = '<div class="result error">Error: ' + error + '</div>';
                });
        }

        function generatePhysical() {
            const count = document.getElementById('physicalCount').value;
            const resultDiv = document.getElementById('physicalResult');
            
            fetch('/api/generate/physical?count=' + count)
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        let html = '<div class="result success">Generated ' + data.inns.length + ' INN(s):</div>';
                        html += '<ul class="inn-list">';
                        data.inns.forEach((inn, index) => {
                            html += '<li>' + (index + 1) + '. ' + inn + '</li>';
                        });
                        html += '</ul>';
                        resultDiv.innerHTML = html;
                    } else {
                        resultDiv.innerHTML = '<div class="result error">' + data.message + '</div>';
                    }
                })
                .catch(error => {
                    resultDiv.innerHTML = '<div class="result error">Error: ' + error + '</div>';
                });
        }

        function generateJuridical() {
            const count = document.getElementById('juridicalCount').value;
            const resultDiv = document.getElementById('juridicalResult');
            
            fetch('/api/generate/juridical?count=' + count)
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        let html = '<div class="result success">Generated ' + data.inns.length + ' INN(s):</div>';
                        html += '<ul class="inn-list">';
                        data.inns.forEach((inn, index) => {
                            html += '<li>' + (index + 1) + '. ' + inn + '</li>';
                        });
                        html += '</ul>';
                        resultDiv.innerHTML = html;
                    } else {
                        resultDiv.innerHTML = '<div class="result error">' + data.message + '</div>';
                    }
                })
                .catch(error => {
                    resultDiv.innerHTML = '<div class="result error">Error: ' + error + '</div>';
                });
        }
    </script>
</body>
</html>
`

// startWebServer starts the web server
func startWebServer(addr string) error {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/validate", handleValidate)
	http.HandleFunc("/api/generate/physical", handleGeneratePhysical)
	http.HandleFunc("/api/generate/juridical", handleGenerateJuridical)

	fmt.Printf("Web server running at http://%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// handleIndex serves the main HTML page
func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// handleValidate handles INN validation requests
func handleValidate(w http.ResponseWriter, r *http.Request) {
	inn := r.URL.Query().Get("inn")
	if inn == "" {
		respondJSON(w, Response{Success: false, Message: "INN parameter is required"})
		return
	}

	valid := ValidateINN(inn)
	respondJSON(w, Response{Success: true, Valid: valid})
}

// handleGeneratePhysical handles physical person INN generation
func handleGeneratePhysical(w http.ResponseWriter, r *http.Request) {
	countStr := r.URL.Query().Get("count")
	count := 5
	if countStr != "" {
		var err error
		count, err = strconv.Atoi(countStr)
		if err != nil || count < 1 || count > 100 {
			respondJSON(w, Response{Success: false, Message: "Invalid count parameter (must be between 1 and 100)"})
			return
		}
	}

	// Additional safety check before allocation
	if count < 1 || count > 100 {
		respondJSON(w, Response{Success: false, Message: "Invalid count parameter (must be between 1 and 100)"})
		return
	}

	inns := make([]string, count)
	for i := 0; i < count; i++ {
		inns[i] = GeneratePhysicalINN()
	}

	respondJSON(w, Response{Success: true, INNs: inns})
}

// handleGenerateJuridical handles juridical person INN generation
func handleGenerateJuridical(w http.ResponseWriter, r *http.Request) {
	countStr := r.URL.Query().Get("count")
	count := 5
	if countStr != "" {
		var err error
		count, err = strconv.Atoi(countStr)
		if err != nil || count < 1 || count > 100 {
			respondJSON(w, Response{Success: false, Message: "Invalid count parameter (must be between 1 and 100)"})
			return
		}
	}

	// Additional safety check before allocation
	if count < 1 || count > 100 {
		respondJSON(w, Response{Success: false, Message: "Invalid count parameter (must be between 1 and 100)"})
		return
	}

	inns := make([]string, count)
	for i := 0; i < count; i++ {
		inns[i] = GenerateJuridicalINN()
	}

	respondJSON(w, Response{Success: true, INNs: inns})
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, response Response) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
