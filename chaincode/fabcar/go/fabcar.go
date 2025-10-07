package main
import (
	"encoding/json"
	"fmt"
	"strconv"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

type SmartContract struct {
}

type Student struct {
	School string `json:"school"`
	Major  string `json:"major"`
	Id     int	`json:"id"`
	Name   string `json:"name"`
}

type Grade struct {
	Course_name string  `json:"course"`
	Course_id   string  `json:"courseId"`
	Teacher     string  `json:"teacher"`
	School      string  `json:"school"`
	Student_id  int     `json:"studentId"`
	Year	    int     `json:"year"`
	Semester    int     `json:"semester"`
	Score       float64 `json:"score"`
}

type Price struct {
	Name        string `json:"name"`
	Id          string `json:"id"`
	Year        int    `json:"year"`
	Level       string `json:"level"`
	Institution string `json:"institution"`
}

func atoi(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return i
}

func atof(str string) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0.0
	}
	return f
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	if function == "addStudent" {
		return s.addStudent(APIstub, args)
	} else if function == "queryStudent" {
		return s.queryStudent(APIstub, args)
	} else if function == "addGrade" {
		return s.addGrade(APIstub, args)
	} else if function == "queryGrade" {
		return s.queryGrade(APIstub, args)
	} else if function == "addPrice" {
		return s.addPrice(APIstub, args)
	} else if function == "queryPrice" {
		return s.queryPrice(APIstub, args)
	}
	return shim.Error("Invalid Smart Contract function name.")
}

// Add a new student
func (s *SmartContract) addStudent(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	var student = Student{School: args[0], Major: args[1], Id: atoi(args[2]), Name: args[3]}
	
	studentAsBytes, _ := json.Marshal(student)
	// studentid+name as key
	err := APIstub.PutState(args[0]+args[2], studentAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to record %s student: %s", args[0], args[2]))
	}

	return shim.Success(nil)
}

// Query a student by school and id
func (s *SmartContract) queryStudent(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	studentAsBytes, _ := APIstub.GetState(args[0]+args[1])
	if studentAsBytes == nil {
		return shim.Error("Could not locate student, the information may not exist")
	}
	return shim.Success(studentAsBytes)
}	

// Add a new grade record
func (s *SmartContract) addGrade(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 8 {
		return shim.Error("Incorrect number of arguments. Expecting 8")
	}

	var grade = Grade{Course_name: args[0], 
					  Course_id: args[1], 
					  Teacher: args[2], 
					  School: args[3], 
					  Student_id: atoi(args[4]), 
					  Year: atoi(args[5]), 
					  Score: atof(args[6]),
					  Semester: atoi(args[7])}
	
	gradeAsBytes, _ := json.Marshal(grade)
	// school+studentid+courseid+year+semester as key
	err := APIstub.PutState(args[3]+args[4]+args[1]+args[5]+args[7], gradeAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to record grade: %s", args[4]))
	}
	return shim.Success(nil)
}

// Query a grade record by school, student id, course id, year and semester
func (s *SmartContract) queryGrade(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	gradeAsBytes, _ := APIstub.GetState(args[0]+args[1]+args[2]+args[3]+args[4])
	if gradeAsBytes == nil {
		return shim.Error("Could not locate grade record, the information may not exist")
	}
	return shim.Success(gradeAsBytes)
}

// Add a new price record
func (s *SmartContract) addPrice(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var price = Price{Name: args[0], 
					  Id: args[1], 
					  Year: atoi(args[2]), 
					  Level: args[3], 
					  Institution: args[4]}
	
	priceAsBytes, _ := json.Marshal(price)
	// Id as key
	err := APIstub.PutState(args[1], priceAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to record price: %s", args[1]))
	}
	return shim.Success(nil)
}

// Query a price record by id
func (s *SmartContract) queryPrice(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	priceAsBytes, _ := APIstub.GetState(args[0])
	if priceAsBytes == nil {
		return shim.Error("Could not locate price record, the information may not exist")
	}
	return shim.Success(priceAsBytes)
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}