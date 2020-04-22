package main

import (
    "fmt"
    "bufio"
    "os"
    "strings"
    "image/jpeg"
    "image/color"
    "image/draw"
    "image/bmp"
    "image"
    "math"
)

func main() {
    reader := bufio.NewReader(os.Stdin)

    fmt.Println("--- Gluck Vtorichny's secret message tool ---")
    fmt.Print("Where is your image (.jpg or .bmp)?\n> ")

    path, err := reader.ReadString('\n')
    check(err)
    path = strings.TrimSpace(path)

    //Open img
    file, err := os.Open(path)
    check(err)

    e := strings.Split(path, ".")
    extension := e[len(e)-1]

    var img image.Image //image.Image interface
    var imgMut *image.RGBA //*image.RGBA implements image.Image

    switch extension {
    case "jpg", "jpeg":
        //Decode jpg
        img, err = jpeg.Decode(file)
        check(err)

        rect := img.Bounds()

        //Convert YCbCr to RGBA
        imgMut = image.NewRGBA(rect)
        draw.Draw(imgMut, rect, img, rect.Min, draw.Src)

        fmt.Printf("Max message length: %d chars\n", rect.Max.X * rect.Max.Y * 3 / 8 - 1)
        fmt.Print("What is your message?\n> ")
        message, _ := reader.ReadString('\n')
        message = strings.TrimSpace(message)

        fmt.Printf("Encoding: %s\n", path)
        encode(imgMut, message)
    case "bmp":
        fmt.Printf("Decoding: %s\n", path)
        //Decode bmp
        img, err = bmp.Decode(file)
        check(err)
        imgMut = img.(*image.RGBA)

        decode(imgMut)
    default:
        fmt.Printf("Unknown extesion: %s", extension)
        os.Exit(1)
    }

    fmt.Println("Done!")
}

func encode(imgMut *image.RGBA, message string) {
    rect := imgMut.Bounds()
    //Check length
    if (rect.Max.X * rect.Max.Y * 3 / 8 - 1) < len(message) {
        fmt.Println("Your image is too small")
    }

    bytes := []byte(message)
    bytes = append(bytes, 0x00) //Null terminated string

    for i := 0; i < len(bytes); i++ {
        bits := byteToBitsReverse(bytes[i])

        for index, val := range bits {
            x, y, n := calculateCoords(i * 8 + index + 1, rect.Max.X, rect.Max.Y)
            imgcolor := imgMut.At(x, y).(color.RGBA)
            switch n {
            case 1:
                imgcolor.R = imgcolor.R & (0xFE | val)
            case 2:
                imgcolor.G = imgcolor.G & (0xFE | val)
            case 3:
                imgcolor.B = imgcolor.B & (0xFE | val)
            }

            imgMut.Set(x, y, imgcolor)
        }
    }

    output, err := os.Create("result.bmp")
    check(err)
    err = bmp.Encode(output, imgMut)
    check(err)
}

func decode(imgMut *image.RGBA) {
    rect := imgMut.Bounds()
    message := make([]byte, 0)

    var buffer uint8

    i := 0
    for {
        i++
        x, y, n := calculateCoords(i, rect.Max.X, rect.Max.Y)
        imgcolor := imgMut.At(x, y).(color.RGBA)
        //Set buffer bit
        switch n {
        case 1:
            buffer = buffer | ((imgcolor.R & 0x01) << uint8((i-1) % 8))
        case 2:
            buffer = buffer | ((imgcolor.G & 0x01) << uint8((i-1) % 8))
        case 3:
            buffer = buffer | ((imgcolor.B & 0x01) << uint8((i-1) % 8))
        }

        //1 byte has been read
        if i % 8 == 0 {
            if buffer == 0x00 {
                break
            }
            message = append(message, byte(buffer))
            buffer = 0
        }

        if i >= rect.Max.X * rect.Max.Y * 3 * 8 {
            break
        }
    }

    fmt.Printf("Message: %s\n", string(message))
}

func check(err error) {
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

//Converts byte to bits (slice) in reversed order
//Returns
//Slice of bits as []uint8
func byteToBitsReverse(b byte) []uint8 {
    bits := make([]uint8, 8)
    for i := 0; i < 8; i++ {
        switch byte(math.Pow(2, float64(i))) & b {
        case 0:
            bits[i] = 0
        default:
            bits[i] = 1
        }
    }

    return bits
}

//Calculate coords of <num>th bit.
//Returns
//x, y - coords of the pixel
//n - 1, 2 or 3 i.e R, G or B byte
func calculateCoords(num, width, height int) (x, y, n int) {
    if (num / 8 / 3) > (width * height) {
        panic("Out of bounds")
    }
    pixelNumber := int(math.Ceil(float64(num) / 3.0))
    //RGB(byte, byte, byte) -> n is 1, 2 or 3
    n = (num-1) % 3 + 1
    x = (pixelNumber-1) % width + 1
    y = int(math.Ceil(float64(pixelNumber) / float64(width)))
    return
}
