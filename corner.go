package main

import (
    "ini"
    "fmt"
    "os"
 fp "path/filepath"
    "path"
    "flag"
    "strconv"
  r "regexp"
  s "strings"
    "io/ioutil"
    "time"
    "gd"
)

// проверяем файл на существование
func fileexists(name string) bool {
    _, e := os.Stat(name)
    return e == nil
}

// проверим как называется ini-файл, если он есть
func getininame() *string {
    if ininame := fp.Base(os.Args[0]) + ".ini"; fileexists(ininame) {
        return &ininame
    }

    if ininame := "makecorner.ini"; fileexists(ininame) {
        return &ininame
    }

    return nil
}

// Универсальное хранилище для ключа командной строки
type option struct {
    kind byte
    k0  *bool
    k1  *int
    k2  *string
    def  string
}

// Константы, описывающие какой тип у содержимого
// в хранилище ключа командной строки
const (
    optionEmpty = iota
    optionBool
    optionInt
    optionString
)

/* Методы хранилища для ключа командной строки */

// установка значение: само значение типа int, bool или string
// и его значение по-умолчанию
func (f *option) Set(value interface{}, def string) {
    switch value.(type) {
        case *bool:
            f.kind = optionBool
            f.k0 = value.(*bool)
        case *int:
            f.kind = optionInt
            f.k1 = value.(*int)
        case *string:
            f.kind = optionString
            f.k2 = value.(*string)
    }

    f.def = def
}

// проверка — является ли значение ключа значением по-умолчанию
func (f option) IsDefault() bool {
    return f.String() == f.def
}

// получить текстовое представление ключа
func (f option) String() string {
    switch f.kind {
        case optionEmpty:
            return "<nil>"

        case optionBool:
            return (map[bool]string{true: "1", false: "0"})[*f.k0]

        case optionInt:
            return strconv.Itoa(*f.k1)

        case optionString:
            return *f.k2
    }

    return ""
}

// функция для разбора опций — из командной строки, заданных по-умолчанию и
// из ini-файла
func parseoptions() map[string]string {
    type optdict struct {
        short string
        def   interface{}
        desc  string
    }

    // массив для обработки ключей:
    // длинная опция, которая опция, знач. по-умолчанию
    def := map[string]optdict {
        "quality":      optdict{"q", 85,    "Качество картинки"},
        "width":        optdict{"w", 660,   "Ширина, высота изменится пропорционально"},
        "radius":       optdict{"r", 10,    "Радиус скругления"},
        "background":   optdict{"b", "#ffffff", "Цвет фона скругления"},
        "mask":         optdict{"m", "*.{j,J}{p,P}{g,G}", "Маска файлов"},
        "out-dir":      optdict{"o", "out", "Папка, куда складываем результат"},
        "save-exif":    optdict{"e", false, "Сохранять ли EXIF"},
        "recursive":    optdict{"R", false, "Рекурсивная обработка"},
        "keep-name":    optdict{"k", false, "Сохранить имена файлов"},
        "moo":          optdict{"M", false, "Му-у-у-у"},
    }

    // Помощь, выводится, если опции заданы неверно или задана опция --help
    flag.Usage = func() {
        fmt.Fprintln(os.Stderr, "Ключи программы:\n")

        for long, optdata := range def {
            var defstr string

            // пропускаем «пасхальное яйцо»
            if long == "moo" {
                continue
            }

            switch optdata.def.(type) {
                case int:
                    defstr = strconv.Itoa(optdata.def.(int))
                case bool:
                    defstr = map[bool]string{true: "установлен", false: "не установлен"}[optdata.def.(bool)]
                case string:
                    defstr = "«" + optdata.def.(string) + "»"
            }


            fmt.Fprintf(os.Stderr, "-%s, --%s — %s (по-умолчанию: %s)\n", optdata.short, long, optdata.desc, defstr)
        }
    }

    // опции из ini-файла
    iniopts := map[string]string{}

    if ininame := getininame(); ininame != nil {
        inimap := ini.ParseFile(*ininame)

        if _, ok := inimap["options"]; ok {
            iniopts = inimap["options"]
        } else {
            iniopts = inimap[""]
        }
    }

    // опции командной строки
    comopts := map[string]option{}

    // Приходится копить опции в указателях, так как настоящие значения из
    // из коммандной строки появятся в переменных только после вызова Parse,
    // который надо делать в самом конце
    for long, optdata := range def {
        for _, name := range [...]string{long, optdata.short} {
            o := option{}
            comopts[name] = o

            switch optdata.def.(type) {
                case int:
                    o.Set(flag.Int(name, optdata.def.(int), optdata.desc), strconv.Itoa(optdata.def.(int)))

                case bool:
                    o.Set(flag.Bool(name, optdata.def.(bool), optdata.desc), (map[bool]string{true: "1", false: "0"})[optdata.def.(bool)])

                case string:
                    o.Set(flag.String(name, optdata.def.(string), optdata.desc), optdata.def.(string))
            }

            // язык не позволяет работать напрямую, приходится через промежуточную переменную
            comopts[name] = o
        }
    }

    flag.Parse()

    // сборка опций из нескольких источников
    options := map[string]string{}

    for long, optdata := range def {
        _, oklong := iniopts[long]
        _, okshrt := iniopts[optdata.short]

        switch {
            case !comopts[long].IsDefault():
                options[long] = comopts[long].String()

            case !comopts[optdata.short].IsDefault():
                options[long] = comopts[optdata.short].String()

            case oklong:
                options[long] = iniopts[long]

            case okshrt:
                options[long] = iniopts[optdata.short]

            default:
                options[long] = comopts[long].String() // если ключ не установлен, в нём значение по-умолчанию
        }
    }

    return options
}

// Сила му-у-у-у-у!
func moo() {
    moo := `
                (__)
                (oo)
           /-----\/
          / |   ||
        *  /\--/\
           ~~  ~~
        `

    os.Stdout.WriteString(moo)
    os.Exit(0)
}

// функция обходит дерево файлов (от текущего местоположения)
// рекурсивно и собирает все файлы, подходящие под маску,
// в процессе обхода исключается папка «except»
func getrecurlist(mask, except string) (out []string) {
    wd, _ := os.Getwd()
    out, _ = fp.Glob(mask)

    for i, file := range out {
        out[i] = fp.Join(wd, file)
    }

    files, e := ioutil.ReadDir(".")
    if e == nil {
        for _, file := range files {
            if file.IsDirectory() {
                entry := path.Join(wd, file.Name)

                if entry == except {
                    continue
                }

                os.Chdir(entry)
                out = append(out, getrecurlist(mask, except)...)
                os.Chdir(wd)
            }
        }
    }

    return
}

func main() {
    options := parseoptions()

    // Показываем силу му-у-у-у-у?
    if v, ok := options["moo"]; ok && v == "1" {
        moo()
    }

    // Преобразование маски файлов в более традиционный для Go формат
    regexp, _ := r.Compile(`(\{[^\}]+\})`)

    oMask := regexp.ReplaceAllStringFunc(options["mask"], func(m string) string {
        return "[" + s.Join(s.Split(m[1:len(m)-1], ",", -1), "") + "]"
    })

    // Первоначальное значение out-dir
    wd, _ := os.Getwd()
    oOutDir := path.Clean(path.Join(wd, options["out-dir"]))

    // Составляем список файлов, по которому будем двигаться
    var oFileList []string

    if options["recursive"] == "1" {
        if path.IsAbs(options["out-dir"]) {
            oOutDir = path.Clean(options["out-dir"])
        }

        oFileList = getrecurlist(oMask, oOutDir)
    } else {
        oFileList, _ = fp.Glob(oMask)
    }

    // Создаём директорий для результата, если он нужен
    if  options["keep-name"] == "0" && !fileexists(oOutDir) {
        os.MkdirAll(oOutDir, 0777)
    }

    // Сколько файлов получилось?
    oLen := len(oFileList)

    if oLen < 1 {
        os.Stdout.WriteString("Файлы не найдены\n")
        os.Exit(1)
    }

    // Маска для нового имени
    now := time.LocalTime().Format("2006.01.02")
    oNameMask := path.Join(options["out-dir"], now)
    if oLen > 1 {
        prec := strconv.Itoa(len(strconv.Itoa(oLen)))
        oNameMask += ".%0" + prec + "d.jpg"
    } else {
        oNameMask += ".jpg"
    }

    // Нормализация background, должны получиться три hex числа
    // в Go очень примитивные regexp :(
    if re := r.MustCompile(`^#[0-9a-fA-F]+`); len(options["background"]) == 7 && !re.MatchString(options["background"]) {
        options["background"] = "#ffffff"
    }

    // И переводим background компоненты
    oBgColor := [3]int{}

    for i := 1; i<len(options["background"]); i+=2 {
        c, _ := strconv.Btoi64(options["background"][i:i+2], 16)
        oBgColor[i >> 1] = int(c)
    }

    // Уголки для скруглений
    var corner *gd.Image
    defer corner.Destroy()

    oRadius, _ := strconv.Atoi(options["radius"])

    if oRadius > 0 {
        corner = gd.CreateTrueColor(oRadius * 2 + 1,  oRadius * 2 + 1)
        corner.AlphaBlending(false)
        corner.SaveAlpha(true)
        trans := corner.ColorAllocateAlpha(0, 0, 0, 127)
        back  := corner.ColorAllocate(oBgColor[0], oBgColor[1], oBgColor[2])

        corner.Fill(0, 0, back)
        corner.FilledEllipse(oRadius, oRadius, oRadius << 1, oRadius << 1, trans)
        corner.GaussianBlur()
    }

    // Качество сохраняемой картинки
    oQuality, _ := strconv.Atoi(options["quality"])

    for num, name := range oFileList {
        im := gd.CreateFromJpeg(name)
        im.AlphaBlending(true)

        sx := im.Sx()
        sy := im.Sy()
        var w, h int

        // Если указана какая-то разумная ширина, то уменьшим до этой
        // ширины
        if w, _ = strconv.Atoi(options["width"]); w > 0 {
            h = int(float32(sy) * (float32(w) / float32(sx)))
            imresized := gd.CreateTrueColor(w, h)
            im.CopyResampled(imresized, 0, 0, 0, 0, w, h, sx, sy)
            im.Destroy()

            im = imresized
        } else {
            w, h = sx, sy
        }

        if R := oRadius; R > 0 {
            corner.Copy(im, 0, 0, 0, 0, R, R)
            corner.Copy(im, 0, h - R, 0, R + 1, R, R)
            corner.Copy(im, w - R, 0, R + 1, 0, R, R)
            corner.Copy(im, w - R, h - R, R + 1, R + 1, R, R)
        }

        if options["keep-name"] == "0" {
            if len(oFileList) > 1 {
                name = fmt.Sprintf(oNameMask, num + 1)
            } else {
                name = oNameMask
            }
        }

        im.Jpeg(name, oQuality)
        im.Destroy()
    }
}
