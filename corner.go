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
        "moo":          optdict{" ", false, "Му-у-у-у"},
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
        for _, name := range []string{long, optdata.short} {
            o := option{}

            switch optdata.def.(type) {
                case int:
                    o.Set(flag.Int(name, optdata.def.(int), optdata.desc), strconv.Itoa(optdata.def.(int)))

                case bool:
                    o.Set(flag.Bool(name, optdata.def.(bool), optdata.desc), (map[bool]string{true: "1", false: "0"})[optdata.def.(bool)])

                case string:
                    o.Set(flag.String(name, optdata.def.(string), optdata.desc), optdata.def.(string))
            }

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
                options[long] = comopts[long].String()
        }
    }

    comopts, comopts = nil, nil

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

func main() {
    options := parseoptions()
    if _, ok := options["moo"]; ok {
        moo()
    }

    fmt.Println(options)
}
