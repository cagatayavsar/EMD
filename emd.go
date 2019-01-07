package main

import (
    "image"
    "image/color"
    "image/png"
    "os"
    "fmt"
    "math"
    "flag"
)

func readImage(path string) image.Image {
    infile, err := os.Open(path)
    if err != nil {
        panic(err.Error())
    }
    defer infile.Close()

    src, _, err := image.Decode(infile)
    if err != nil {
        panic(err.Error())
    }
    return src 
}

func rgbToGray(im image.Image) *image.Gray {

    bounds := im.Bounds()
    width, height := bounds.Max.X, bounds.Max.Y
    gray := image.NewGray(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})

    for x := 0; x < width; x++ {
        for y := 0; y < height; y++ {
            
            oldColor := im.At(x, y) 
            r, g, b, _ := oldColor.RGBA()            
            avg := (float64(r) + float64(g) + float64(b)) / 3
            var grayColor color.Gray
            grayColor = color.Gray{uint8(avg/256)}
            gray.Set(x, y, grayColor)
        }
    }

    return gray
}

func writeImage(im image.Image, imName string) {

    outfile, err := os.Create(imName)
    if err != nil {
        panic(err.Error())
    }
    defer outfile.Close()
    png.Encode(outfile, im) 
}

func reverseArray(numbers []int) []int {
    
    for i, j := 0, len(numbers)-1; i < j; i, j = i + 1, j-1 {
        numbers[i], numbers[j] = numbers[j], numbers[i]
    }
    return numbers
}

func toBase(val int, base int) []int {
    
    // (11,5) -> number 11 to base five
    // it returns [2, 1]
    temp := []int{}
    for true {
        temp = append(temp, val % base)
        if(val >= base) {

            val /= base 
        }else {
            
            break
        }
    }
    return reverseArray(temp)
}

func baseTo(secVal []int, base int) int {

    // it takes array and base of array and return ten base 
    i := 0
    val := 0

    for i = 0; i < len(secVal); i++ {
        val += secVal[i] * int(math.Pow(float64(base),float64(len(secVal) - i - 1)))
    }

    return val
}

func encryption(cover *image.Gray, stego *image.Gray) *image.Gray{
    
    //0-2 pixels for which module (%5 or %7)
    //2-10 pixels for stego image width size with %5 encryption
    //10-18 pixels for stego image height size with %5 encryption
    numOfPixMode := 2
    numOfPixWidth := 8
    numOfPixHeight := 8

    //how many pixels using for header
    numOfPixHeader := numOfPixMode + numOfPixWidth + numOfPixHeight

    //numbers from 0 to 255 can represent 3 digits in module 7
    //x+2y+3z(3 pixel) and 3 digits
    modeSevenPixel := 3
    //9 cover pixels for 1 stego pixel
    numOfPixelModule7 := 3 * modeSevenPixel

    //numbers from 0 to 255 can represent 4 digits in module 5
    //x+2y(2 pixel) and 4 digits
    modeFivePixel := 2
    //8 cover pixels for 1 stego pixel
    numOfPixelModule5 := 4 * modeFivePixel
    
    
    coverBounds := cover.Bounds()
    //cover image width and height
    coverW, coverH := coverBounds.Max.X, coverBounds.Max.Y

    coverAfter := image.NewGray(image.Rectangle{image.Point{0, 0}, image.Point{coverW, coverH}})

    for i := 0; i < coverW * coverH; i++ {

        if(cover.Pix[i] == 255) {
            cover.Pix[i] = 254
        }else if(cover.Pix[i] == 0) {
            cover.Pix[i] = 1
        }
        coverAfter.Pix[i] = cover.Pix[i]
    }

    stegoBounds := stego.Bounds()
    //stego image width and height
    stegoW, stegoH := stegoBounds.Max.X, stegoBounds.Max.Y
    //if mode is 0 then use module 5 for encryption
    //if mode is 1 then use module 7 for encryption
    //if mode is -1, can't encrypt 
    mode := -1

    
    if(coverW * coverH > (stegoW * stegoH * numOfPixelModule7) + numOfPixHeader) {
        //with module 7
        mode = 1
    }else if(coverW * coverH > (stegoW * stegoH * numOfPixelModule5) + numOfPixHeader) {
        //with module 5
        mode = 0
    }
    
    //if there is no problem with encryption we can set header information
    if(mode != -1) {
        
        //first two pixel -> which mode, 0 -> module 5, 1 -> module 7
        val := int(int(coverAfter.Pix[0]) + 2 * int(coverAfter.Pix[1])) % 5
        diff := val - mode

        if (diff == 1) {
            coverAfter.Pix[0] -= 1
        }else if (diff == 2) {
            coverAfter.Pix[1] -= 1
        }else if (diff == 3) {
            coverAfter.Pix[1] += 1
        }else if (diff == -1) || (diff == 4) {
            coverAfter.Pix[0] += 1
        }

        //2-10 pixels are stego image width
        //max width = 624
        //using (x+2y)%5 for hide stego image width
        //we have 8 pixel for hide width
        //x+2y(2 pixel), then we can hide 4 digit 
        secVal := toBase(int(stegoW), 5)
        
        //how many zeros will be added before the number
        numZeros := (numOfPixWidth / modeFivePixel) - len(secVal)

        for j := 0; j < numOfPixWidth / modeFivePixel; j++{
            
            //index starts with 2, after 2 pixels mode
            index := modeFivePixel * j + numOfPixMode
            val := int(int(coverAfter.Pix[index]) + 2 * int(coverAfter.Pix[index + 1])) % 5
            diff := 0
            
            if (j < numZeros) {

                diff = val
            }else {

                diff = val - secVal[j - numZeros]
            }

            if (diff == -4) || (diff == 1) {
                coverAfter.Pix[index] -= 1
            }else if (diff == -3) || (diff == 2) {
                coverAfter.Pix[index+1] -= 1
            }else if (diff == -2) || (diff == 3) {
                coverAfter.Pix[index+1] += 1
            }else if (diff == -1) || (diff == 4) {
                coverAfter.Pix[index] += 1
            }
        }
        
        //10-18 pixels are steho image height
        //max height = 624
        //using (x+2y)%5 for hide stego image height
        //we have 8 pixel for hide height
        //x+2y(2 pixel), then we can hide 4 digit 
        secVal = toBase(int(stegoH), 5)
        
        //how many zeros will be added before the number
        numZeros = (numOfPixHeight / modeFivePixel) - len(secVal)

        for j := 0; j < numOfPixHeight / modeFivePixel; j++{
            
            //index starts with 10, after 2 pixel mode and 8 pixel width
            index := modeFivePixel * j + numOfPixMode + numOfPixWidth
            val := int(int(coverAfter.Pix[index]) + 2 * int(coverAfter.Pix[index + 1])) % 5
            diff := 0
            
            if (j < numZeros) {
                
                diff = val
            }else {
                
                diff = val - secVal[j - numZeros]
            }

            if (diff == -4) || (diff == 1) {
                coverAfter.Pix[index] -= 1
            }else if (diff == -3) || (diff == 2) {
                coverAfter.Pix[index+1] -= 1
            }else if (diff == -2) || (diff == 3) {
                coverAfter.Pix[index+1] += 1
            }else if (diff == -1) || (diff == 4) {
                coverAfter.Pix[index] += 1
            }
        }

    }

    
    if(mode == 1) {
        //(x+2y+3z)%7
        fmt.Println("Module       : 7")

        for i := 0; i < stegoW * stegoH; i++ {

            secVal := toBase(int(stego.Pix[i]), 7)

            //how many zeros will be added before the number
            numZeros := numOfPixelModule7 / modeSevenPixel - len(secVal)

            for j := 0; j < numOfPixelModule7 / modeSevenPixel; j++ {

                //index starts with 18, after 18 pixel header
                index := modeSevenPixel * j + numOfPixelModule7 * i + numOfPixHeader

                val := int(int(coverAfter.Pix[index]) + 2 * int(coverAfter.Pix[index + 1]) + 3 * int(coverAfter.Pix[index + 2])) % 7

                diff := 0

                if (j < numZeros) {
                    
                    diff = val
                }else {
                    
                    diff = val - secVal[j - numZeros]
                }

                if (diff == -6) || (diff == 1) {
                    coverAfter.Pix[index] -= 1
                }else if (diff == -5) || (diff == 2) {
                    coverAfter.Pix[index+1] -= 1
                }else if (diff == -4) || (diff == 3) {
                    coverAfter.Pix[index+2] -= 1
                }else if (diff == -3) || (diff == 4) {
                    coverAfter.Pix[index+2] += 1
                }else if (diff == -2) || (diff == 5) {
                    coverAfter.Pix[index+1] += 1
                }else if (diff == -1) || (diff == 6) {
                    coverAfter.Pix[index] += 1
                }
            }
        } 

    }else if(mode == 0) {
        
        //(x+2y)%5
        fmt.Println("Module 5")
        
        for i := 0; i < stegoW * stegoH; i++ {
            
            secVal := toBase(int(stego.Pix[i]), 5)
            
            //how many zeros will be added before the number
            numZeros := numOfPixelModule5 / modeFivePixel - len(secVal)

            for j := 0; j < numOfPixelModule5 / modeFivePixel; j++ {

                //index starts with 18, after 18 pixel header
                index := modeFivePixel * j + numOfPixelModule5 * i + numOfPixHeader

                val := int(int(coverAfter.Pix[index]) + 2 * int(coverAfter.Pix[index + 1])) % 5

                diff := 0
                if (j < numZeros){
                	diff = val
                }else{
                	diff = val - secVal[j - numZeros]
                }
                
                if (diff == -4) || (diff == 1) {
                    coverAfter.Pix[index] -= 1
                }else if (diff == -3) || (diff == 2) {
                    coverAfter.Pix[index+1] -= 1
                }else if (diff == -2) || (diff == 3) {
                    coverAfter.Pix[index+1] += 1
                }else if (diff == -1) || (diff == 4) {
                    coverAfter.Pix[index] += 1
                }
            }
        }        
    }else if(mode == -1) {
        fmt.Println("Stego width*height must be smaller than", coverH * coverW / 8)
        fmt.Println("Your stego image width:", stegoW, ", height:", stegoH, "->", stegoW * stegoH) 
        fmt.Println("can't encrypted")
    }else {
        
        fmt.Println("something went wrong!")
    }

    return coverAfter
}

func decryption(cover *image.Gray) *image.Gray{

    //0-2 pixels for which module (%5 or %7)
    //2-10 pixels for stego image width size with %5 encryption
    //10-18 pixels for stego image height size with %5 encryption
    numOfPixMode := 2
    numOfPixWidth := 8
    numOfPixHeight := 8

    //how many pixels using for header
    numOfPixHeader := numOfPixMode + numOfPixWidth + numOfPixHeight

    //numbers from 0 to 255 can represent 3 digits in module 7
    //x+2y+3z(3 pixel) and 3 digits
    modeSevenPixel := 3
    //9 cover pixels for 1 stego pixel
    numOfPixelModule7 := 3 * modeSevenPixel

    //numbers from 0 to 255 can represent 4 digits in module 5
    //x+2y(2 pixel) and 4 digits
    modeFivePixel := 2
    //8 cover pixels for 1 stego pixel
    numOfPixelModule5 := 4 * modeFivePixel
    
    //calculate which mode used 0 -> module 5, 1 -> module 7
    mode := (int(cover.Pix[0]) + 2 * int(cover.Pix[1])) % 5
    fmt.Println("mode:", mode)
    //calculate stego image width    
    val := []int{}
    for j := 0; j < numOfPixWidth / modeFivePixel; j++{
        
        index := modeFivePixel * j + numOfPixMode
        val = append(val, (int(cover.Pix[index]) + 2 * int(cover.Pix[index + 1])) % 5)
    }
    width := baseTo(val,5)

    //calculate stego image height
    val = []int{}
    for j := 0; j < numOfPixHeight / modeFivePixel; j++{

        index := modeFivePixel * j + numOfPixMode + numOfPixWidth
        val = append(val, (int(cover.Pix[index]) + 2 * int(cover.Pix[index + 1])) % 5)
    }
    height := baseTo(val,5)

    stego := image.NewGray(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})
    
    if (mode == 0) {
        
        //(x+2y)%5
        for i := 0; i < width * height; i++ {
            
            val := []int{}
            for j := 0; j < numOfPixelModule5 / modeFivePixel; j++ {
                
                index := modeFivePixel * j + numOfPixelModule5 * i + numOfPixHeader
                val = append(val, (int(cover.Pix[index]) + 2 * int(cover.Pix[index + 1])) % 5)
            }
            value := baseTo(val,5)
            stego.Pix[i] = uint8(value)
        }

    }else if (mode == 1){
        
        //(x+2y+3z)%7
        for i := 0; i < width * height; i++ {
            
            val := []int{}
            for j := 0; j < numOfPixelModule7 / modeSevenPixel; j++ {
                
                index := modeSevenPixel * j + numOfPixelModule7 * i + numOfPixHeader
                val = append(val, (int(cover.Pix[index]) + 2 * int(cover.Pix[index + 1]) + 3 * int(cover.Pix[index + 2])) % 7)
            }
            value := baseTo(val,7)
            stego.Pix[i] = uint8(value)
        }

    }else {
        fmt.Println("Something wrong with decryption!")
    }
    
    return stego
}

func psnr(coverBefore *image.Gray, coverAfter *image.Gray) float64{
    
    coverBounds := coverBefore.Bounds()
    //cover image width and height
    coverW, coverH := coverBounds.Max.X, coverBounds.Max.Y

    var mse float64 = 0
    
    for i := 0; i < coverW * coverH; i++ {
        mse += math.Pow(float64(float64(coverBefore.Pix[i]) - float64(coverAfter.Pix[i])), float64(2))
        
    }

    mse = math.Sqrt(mse / float64(coverW * coverH))
    return (20 * math.Log10(254/mse))
}

func main() {
    
    var enc,dec bool
    flag.BoolVar(&enc, "e", false, "Encryption")
    flag.BoolVar(&dec, "d", false, "Decryption")
    
    var coverPath, stegoPath, outputPath string
    flag.StringVar(&coverPath, "c", "", "Cover Image Path")
    flag.StringVar(&stegoPath, "s", "", "Stego Image Path")
    flag.StringVar(&outputPath, "o", "", "Output Image Path")

    flag.Usage = func() {
        fmt.Printf("Usage of %s: [OPTION FILE FILE...]\n", os.Args[0][2:])
        fmt.Printf("\n")

        fmt.Printf("  [OPTION]\n")
        fmt.Printf("  -e boolean Encryption\n")
        fmt.Printf("  -d boolean Decryption\n\n")
        
        fmt.Printf("  [FILE]\n")
        fmt.Printf("  -c string  Cover Image Path\n")
        fmt.Printf("  -s string  Stego Image Path\n")
        fmt.Printf("  -o string  Output Image Path\n\n")
        
        fmt.Printf("  Examples...\n")
        fmt.Printf("  to Encryption  : %s -e -c=Input.png -s=Hidden.png -o=Output.png\n", os.Args[0])
        fmt.Printf("                   %s -e -c Input.png -s Hidden.png -o Output.png\n", os.Args[0])

        fmt.Printf("  to Decryption  : %s -d -c=Input.png -o=Output.png\n", os.Args[0])
    }
    flag.Parse()
    
    if (flag.NFlag() == 0) {
        flag.Usage()
        os.Exit(1)
    }

    if(enc == true && dec == false) {
        
        fmt.Println(".............Encryption.............")
        fmt.Println("Cover Image  :", coverPath)
        fmt.Println("Stego Image  :", stegoPath)
        fmt.Println("Output Image :", outputPath)

        cover := readImage(coverPath)
        stego := readImage(stegoPath)

        coverGray := rgbToGray(cover)
        stegoGray := rgbToGray(stego)

        fmt.Println("Cover Gray   : coverGray.png")
        fmt.Println("Stego Gray   : stegoGray.png")

        coverAfter := encryption(coverGray, stegoGray)
        writeImage(coverAfter, outputPath)
        writeImage(coverGray, "coverGray.png")
        writeImage(stegoGray, "stegoGray.png")

        fmt.Println("PSNR         :", psnr(coverAfter, coverGray))
        fmt.Println(".........End of Encryption..........")

    }else if(enc == false && dec == true) {
        
        fmt.Println(".............Decryption............")
        fmt.Println("Output Image :", outputPath)

        cover := readImage(coverPath)
        coverGray := rgbToGray(cover)

        stego := decryption(coverGray)
        writeImage(stego, outputPath)
        fmt.Println(".........End of Decryption.........")

    }else {
        
        flag.Usage()
        fmt.Printf("\n  Choose one option -e or -d !!!\n")
        os.Exit(1)
    }
}