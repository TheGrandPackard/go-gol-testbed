package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

var fileNameListCRC = 0x61580AC9

func main() {
	// _, _, _, err := loadOBJ("cube.obj")
	// if err != nil {
	// 	log.Printf("%s\n", err.Error())
	// }

	err := loadS3d("befallen.s3d")
	if err != nil {
		log.Printf("%s\n", err.Error())
	}
	// err = loadS3d("befallen_chr.s3d")
	// if err != nil {
	// 	log.Printf("%s\n", err.Error())
	// }
}

func loadOBJ(fileName string) ([]float32, []float32, []float32, error) {
	var vertices, uvs, normals []float32

	objFile, err := os.Open(fileName)
	if err != nil {
		return vertices, uvs, normals, fmt.Errorf("obj file %q not found on disk: %v", fileName, err)
	}
	defer objFile.Close()

	scanner := bufio.NewScanner(objFile)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return vertices, uvs, normals, nil
}

type PFSHeader struct {
	Offset      uint32
	MagicCookie [4]byte
	Unknown     uint32
}

type DirectoryHeader struct {
	Count uint32
}

type ByOffset []FileHeader

func (a ByOffset) Len() int           { return len(a) }
func (a ByOffset) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByOffset) Less(i, j int) bool { return a[i].Offset < a[j].Offset }

type FileHeader struct {
	CRC    uint32
	Offset uint32
	Size   uint32
}

type DataBlock struct {
	CompressedLength uint32
	InflatedLenth    uint32
}

type FileNameCount struct {
	Count uint32
}

type FileNameLength struct {
	Length uint32
}

func loadS3d(fileName string) error {
	fmt.Printf("Loading s3d Archive: %s\n", fileName)

	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("s3d file %q not found on disk: %v", fileName, err)
	}
	defer file.Close()

	headerBytes := make([]byte, 12)
	header := PFSHeader{}
	_, err = file.Read(headerBytes)
	if err != nil {
		return fmt.Errorf("Error reading PFS header bytes: %v", err)
	}
	buffer := bytes.NewBuffer(headerBytes)
	err = binary.Read(buffer, binary.LittleEndian, &header)
	if err != nil {
		return fmt.Errorf("binary.Read failed: %v", err)
	}
	// fmt.Printf("Directory Header Offset: %X\n", header.Offset)

	// Validate header
	if string(header.MagicCookie[:]) != "PFS " {
		return fmt.Errorf("Magic Cookie Not PFS")
	}
	if header.Unknown != 131072 {
		return fmt.Errorf("Unknown Header Value Not Equal To 131072")
	}

	directoryHeaderBytes := make([]byte, 4)
	directoryHeader := DirectoryHeader{}
	_, err = file.ReadAt(directoryHeaderBytes, int64(header.Offset))
	if err != nil {
		return fmt.Errorf("Error reading directory header bytes: %v", err)
	}
	buffer = bytes.NewBuffer(directoryHeaderBytes)
	err = binary.Read(buffer, binary.LittleEndian, &directoryHeader)
	if err != nil {
		return fmt.Errorf("binary.Read failed: %v", err)
	}
	// fmt.Printf("File Count: %d\n", directoryHeader.Count)

	// Get file crcs, offsets, and checksums
	fileHeaders := make([]FileHeader, 0)
	fileNameHeader := FileHeader{}
	for i := 0; i < int(directoryHeader.Count); i++ {
		fileHeaderBytes := make([]byte, 12)
		fileHeader := FileHeader{}
		_, err = file.ReadAt(fileHeaderBytes, int64(int(header.Offset)+4+i*12))
		if err != nil {
			return fmt.Errorf("Error reading file header bytes: %v", err)
		}
		buffer = bytes.NewBuffer(fileHeaderBytes)
		err = binary.Read(buffer, binary.LittleEndian, &fileHeader)
		if err != nil {
			return fmt.Errorf("binary.Read failed: %v", err)
		}
		// fmt.Printf("Parsed data: %+v\n", directoryHeader)

		if fileHeader.CRC == uint32(fileNameListCRC) {
			fileNameHeader = fileHeader
			// fmt.Printf("Directory Header Found. CRC: %X Offset: %X Size: %X\n", fileHeader.CRC, fileHeader.Offset, fileHeader.Size)
		} else {
			fileHeaders = append(fileHeaders, fileHeader)
			// fmt.Printf("File Header Found. CRC: %X Offset: %X Size: %X\n", fileHeader.CRC, fileHeader.Offset, fileHeader.Size)
		}
	}

	// Sort the offsets
	sort.Sort(ByOffset(fileHeaders))

	// Get file names
	fileNames := make([]string, 0)
	fileNameDataBlockBytes := make([]byte, 8)
	fileNameDataBlock := DataBlock{}
	_, err = file.ReadAt(fileNameDataBlockBytes, int64(fileNameHeader.Offset))
	if err != nil {
		return fmt.Errorf("Error reading data block bytes: %v", err)
	}
	buffer = bytes.NewBuffer(fileNameDataBlockBytes)
	err = binary.Read(buffer, binary.LittleEndian, &fileNameDataBlock)
	if err != nil {
		return fmt.Errorf("binary.Read failed: %v", err)
	}
	// TODO: Read more file name blocks if necessary
	// fmt.Printf("Filename Header Block Found. Compressed Length: %X Inflated Length: %X\n", fileNameDataBlock.CompressedLength, fileNameDataBlock.InflatedLenth)

	fileNameBytes := make([]byte, fileNameDataBlock.CompressedLength)
	_, err = file.ReadAt(fileNameBytes, int64(fileNameHeader.Offset+8))
	if err != nil {
		return fmt.Errorf("Error reading file name bytes: %v", err)
	}
	buffer = bytes.NewBuffer(fileNameBytes)
	r, err := zlib.NewReader(buffer)
	if err != nil {
		panic(err)
	}
	r.Close()

	// File count
	fileNameCountBytes := make([]byte, 4)
	fileNameCount := FileNameCount{}
	_, err = r.Read(fileNameCountBytes)
	if err != nil {
		return fmt.Errorf("Error reading file name count bytes: %v", err)
	}
	buffer = bytes.NewBuffer(fileNameCountBytes)
	err = binary.Read(buffer, binary.LittleEndian, &fileNameCount)
	if err != nil {
		return fmt.Errorf("binary.Read failed: %v", err)
	}
	// fmt.Printf("Found %d Files\n", fileNameCount.Count)

	os.Mkdir("output", 0755)
	fileNoExt := strings.Split(fileName, ".")[0]
	os.Mkdir("output/"+fileNoExt, 0755)
	for i := 0; i < int(fileNameCount.Count); i++ {
		// File length
		fileNameLengthBytes := make([]byte, 4)
		fileNameLength := FileNameLength{}
		_, err = r.Read(fileNameLengthBytes)
		if err != nil {
			return fmt.Errorf("Error reading file name length bytes: %v", err)
		}
		buffer = bytes.NewBuffer(fileNameLengthBytes)
		err = binary.Read(buffer, binary.LittleEndian, &fileNameLength)
		if err != nil {
			return fmt.Errorf("binary.Read failed: %v", err)
		}
		// fmt.Printf("Filename %d Length Found: %X bytes\n", i+1, fileNameLength.Length)

		// File name
		fileNameEntryBytes := make([]byte, fileNameLength.Length)
		_, err = r.Read(fileNameEntryBytes)
		if err != nil && err != io.EOF {
			return fmt.Errorf("Error reading file name entry bytes: %v", err)
		}
		fileName := string(bytes.Trim(fileNameEntryBytes, "\x00"))
		fileNames = append(fileNames, fileName)
		fmt.Printf("%X\t%s\t%X\n", fileHeaders[i].Offset, fileNames[i], fileHeaders[i].Size)

		// Extract file
		var inflated uint32
		f, err := os.Create(fmt.Sprintf("output/%s/%s", fileNoExt, fileNames[i]))
		if err != nil {
			return fmt.Errorf("Error opening file to write: %v", err)
		}

		// Read multiple blocks (this only reads a single 8k block and writes it out)
		file.Seek(int64(fileHeaders[i].Offset), 0)
		for inflated < fileHeaders[i].Size {
			fileDataBlockBytes := make([]byte, 8)
			fileDataBlock := DataBlock{}
			_, err = file.Read(fileDataBlockBytes)
			if err != nil {
				return fmt.Errorf("Error reading data block bytes: %v", err)
			}
			buffer = bytes.NewBuffer(fileDataBlockBytes)
			err = binary.Read(buffer, binary.LittleEndian, &fileDataBlock)
			if err != nil {
				return fmt.Errorf("binary.Read failed: %v", err)
			}
			// fmt.Printf("Data Block Compressed: %X Inflated: %X\n", fileDataBlock.CompressedLength, fileDataBlock.InflatedLenth)
			fileBytes := make([]byte, fileDataBlock.CompressedLength)
			_, err = file.Read(fileBytes)
			if err != nil {
				return fmt.Errorf("Error reading file name bytes: %v", err)
			}
			buffer = bytes.NewBuffer(fileBytes)
			r2, err := zlib.NewReader(buffer)
			if err != nil {
				panic(err)
			}
			io.Copy(f, r2)
			r2.Close()

			inflated += fileDataBlock.InflatedLenth
		}

		f.Close()
	}

	return nil
}
