package epedit

import (
	"bufio"
	"os"
)

// IDD에서 정의하는 단일 필드의 "규칙" (예: 필수 여부, 타입)
type FieldDef struct {
	Name             string
	Required         bool
	Units            string
	Default          string // 일단 값을 string으로, 타입 변환은 나중에 생각...
	Autosizable      bool
	Autocalculatable bool
	Type             string   // alpha, real, integer, choice 등
	Choices          []string // type이 choice인 경우 해당
	// ... 필요에 따라 Minimum, Maximum 등 추가
}

// IDD에서 \extensible 속성을 가진 클래스의 확장 규칙을 정의
type ExtensibleDef struct {
	BeginIndex int // extensible 필드가 시작되는 인덱스
	Size       int // extensible 필드 크기 (ex. X, Y, Z 좌표가 반복되면 3)
}

// IDD에서 정의하는 하나의 오브젝트 "클래스 스키마" (예: Building, Zone)
type ClassDef struct {
	Name       string     // 대소문자가 구분된 원래 이름
	Group      string     // \group
	Fields     []FieldDef // 규칙들의 배열
	MinFields  int
	Extensible *ExtensibleDef // 없을 때 nil이 됨
}

// 전체 스키마를 담고 있는 사전 오브젝트
type IDD struct {
	Version string
	Classes map[string]*ClassDef // 대소문자 구분 없이 빠른 검색을 위한 Map
}

func NewIDD() *IDD {
	return &IDD{
		Classes: make(map[string]*ClassDef),
	}
}

// filepath의 IDD 파일을 읽어 파싱된 IDD 구조체의 포인터를 반환
func ParseIDD(filepath string) (*IDD, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	idd := NewIDD()
	scanner := bufio.NewScanner(file)

	return idd, err
}
