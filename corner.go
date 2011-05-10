package main

import (
    "ini"
    "fmt"
    "os"
 fp "path/filepath"
    "flag"
    "strconv"
)

// проверяем файл на существование
func fileexists(name string) bool {
    fi, e := os.Stat(name)

    return e == nil && fi.IsRegular()
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
            if *f.k0 {
                return "1"
            }
            return "0"

        case optionInt:
            return strconv.Itoa(*f.k1)

        case optionString:
            return *f.k2
    }

    return ""
}

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
    }

    flag.Usage = func() {
        fmt.Fprintln(os.Stderr, "Ключи программы:\n")

        for long, optdata := range def {
            var defstr string

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

    for long, optdata := range def {
        optlong, optshrt := option{}, option{}

        switch optdata.def.(type) {
            case int:
                optlong.Set(flag.Int(long, optdata.def.(int), optdata.desc), strconv.Itoa(optdata.def.(int)))
                optshrt.Set(flag.Int(optdata.short, optdata.def.(int), optdata.desc), strconv.Itoa(optdata.def.(int)))

                comopts[long], comopts[optdata.short] = optlong, optshrt

            case bool:
                defstr := map[bool]string{true: "1", false: "0"}

                optlong.Set(flag.Bool(long, optdata.def.(bool), optdata.desc), defstr[optdata.def.(bool)])
                optshrt.Set(flag.Bool(optdata.short, optdata.def.(bool), optdata.desc), defstr[optdata.def.(bool)])

                comopts[long], comopts[optdata.short] = optlong, optshrt

            case string:
                optlong.Set(flag.String(long, optdata.def.(string), optdata.desc), optdata.def.(string))
                optshrt.Set(flag.String(optdata.short, optdata.def.(string), optdata.desc), optdata.def.(string))

                comopts[long], comopts[optdata.short] = optlong, optshrt
       }
    }

    flag.Parse()
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
                options[long] = comopts[long].String()
        }
    }

    comopts, comopts = nil, nil

    return options
}

func main() {
    options := parseoptions()

    fmt.Println(options)
}
