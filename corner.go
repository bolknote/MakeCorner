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
	"jpegtran"
	"exec"
	"runtime"
	"getncpu"
	"math"
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
	k0   *bool
	k1   *int
	k2   *string
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
	def := map[string]optdict{
		"quality":    optdict{"q", 85, "Качество картинки"},
		"width":      optdict{"w", 660, "Ширина, высота изменится пропорционально"},
		"radius":     optdict{"r", 10, "Радиус скругления"},
		"background": optdict{"b", "#ffffff", "Цвет фона скругления"},
		"mask":       optdict{"m", "*.{j,J}{p,P}{g,G}", "Маска файлов"},
		"out-dir":    optdict{"o", "out", "Папка, куда складываем результат"},
		"save-exif":  optdict{"e", false, "Сохранять ли EXIF"},
		"recursive":  optdict{"R", false, "Рекурсивная обработка"},
		"keep-name":  optdict{"k", false, "Сохранить имена файлов"},
		"moo":        optdict{"M", false, "Му-у-у-у"},
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

// Определяем — это серое изображение?
func isgray(im *gd.Image) bool {
	for x := im.Sx() - 1; x >= 0; x-- {
		for y := im.Sy() - 1; y >= 0; y-- {
			c := im.ColorsForIndex(im.ColorAt(x, y))

			if c["red"] != c["green"] || c["red"] != c["blue"] {
				return false
			}
		}
	}

	return true
}

// функция грубой отрисовки части дуги окружности
// по алгоритму Ulrich Mierendorf (imageSmoothArc_optimized)
func smootharc(p *gd.Image, cx, cy, a, b float64, fillColor gd.Color, start, stop, seg float64) {
	color := p.ColorsForIndex(fillColor)
	var xp, yp, xa, ya float64

	switch seg {
	case 0:
		xp, yp, xa, ya = 1, -1, 1, -1
	case 1:
		xp, yp, xa, ya = -1, -1, 0, -1
	case 2:
		xp, yp, xa, ya = -1, 1, 0, 0
	case 3:
		xp, yp, xa, ya = 1, 1, 1, 0
	}

	for x := float64(0); x <= a; x++ {
		y := b * math.Sqrt(1-(x*x)/(a*a))
		error := y - float64(int(y))
		y = float64(int(y))

		alpha := int(127 - float64(127-color["alpha"])*error)
		diffColor := p.ColorExactAlpha(color["red"], color["green"], color["blue"], alpha)

		xx := int(cx + xp*x + xa)

		p.SetPixel(xx, int(cy+yp*(y+1)+ya), diffColor)
		p.Line(xx, int(cy+yp*y+ya), xx, int(cy+ya), fillColor)
	}

	for y := float64(0); y < b; y++ {
		x := a * math.Sqrt(1-(y*y)/(b*b))
		error := x - float64(int(x))
		x = float64(int(x))

		alpha := int(127 - float64(127-color["alpha"])*error)
		diffColor := p.ColorExactAlpha(color["red"], color["green"], color["blue"], alpha)
		p.SetPixel(int(cx+xp*(x+1)+xa), int(cy+yp*y+ya), diffColor)
	}
}

// округление
func round(f float64) float64 {
	if f-float64(int(f)) >= 0.5 {
		return math.Ceil(f)
	}

	return math.Floor(f)
}


// Эллипс с антиалиасингом
// по алгоритму Ulrich Mierendorf (imageSmoothArc_optimized)
func smoothellipse(p *gd.Image, cx, cy, r int, c gd.Color) {
	for i := float64(0); i < 4; i++ {
		stop := (i + 1) * math.Pi / 2

		if stop/2 < math.Pi*2 {
			smootharc(p, float64(cx), float64(cy), float64(r), float64(r), c, 0, stop, i)
		} else {
			smootharc(p, float64(cx), float64(cy), float64(r), float64(r), c, 0, math.Pi*2, i)
			break
		}
	}
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
	if options["keep-name"] == "0" && !fileexists(oOutDir) {
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

		fmt.Printf("Found %d JPEG files.\n", oLen)
	} else {
		oNameMask += ".jpg"
		fmt.Println("Found 1 JPEG file.")
	}

	// Нормализация background, должны получиться три hex числа
	// в Go очень примитивные regexp :(
	if re := r.MustCompile(`^#[0-9a-fA-F]+`); len(options["background"]) == 7 && !re.MatchString(options["background"]) {
		options["background"] = "#ffffff"
	}

	// И переводим background компоненты
	oBgColor := [3]int{}

	for i := 1; i < len(options["background"]); i += 2 {
		c, _ := strconv.Btoi64(options["background"][i:i+2], 16)
		oBgColor[i>>1] = int(c)
	}

	// Уголки для скруглений
	var corner *gd.Image
	defer corner.Destroy()

	oRadius, _ := strconv.Atoi(options["radius"])

	if oRadius > 0 {
		corner = gd.CreateTrueColor(oRadius<<1+2, oRadius<<1+2)
		corner.AlphaBlending(false)
		corner.SaveAlpha(true)
		trans := corner.ColorAllocateAlpha(oBgColor[0], oBgColor[1], oBgColor[2], 127)
		back := corner.ColorAllocate(oBgColor[0], oBgColor[1], oBgColor[2])

		corner.Fill(0, 0, trans)
		//corner.SmoothFilledEllipse(oRadius, oRadius, oRadius << 1, oRadius << 1, back)
		smoothellipse(corner, oRadius, oRadius+1, oRadius, back)

		// инвертируем прозрачность пикселей
		for x := 0; x < corner.Sx(); x++ {
			for y := 0; y < corner.Sy(); y++ {
				c := corner.ColorsForIndex(corner.ColorAt(x, y))
				c["alpha"] = 127 - c["alpha"]

				nc := corner.ColorAllocateAlpha(c["red"], c["green"], c["blue"], c["alpha"])
				corner.SetPixel(x, y, nc)
			}
		}
	}

	// Качество сохраняемой картинки
	oQuality, _ := strconv.Atoi(options["quality"])

	// Это временное имя для утилиты jpegtran, которую распаковываем из архива
	jtexe := fp.Join(os.TempDir(), "corner-bolk-jpegtran.exe")
	defer os.Remove(jtexe)

	// Временное имя для ч/б профиля
	oProfile := fp.Join(os.TempDir(), "cornet-bolk-bw.txt")
	defer os.Remove(oProfile)

	// Выставляем количество процессов, которые могут выполняться одновременно
	// равное количеству процессоров
	if n := getncpu.Getncpu(); n > 0 {
		runtime.GOMAXPROCS(n)
	}

	// Выводим сколько процессоров мы намерены использовать
	{
		n := runtime.GOMAXPROCS(-1)
		if n == 1 {
			fmt.Println("1 CPU will be use.")
		} else {
			fmt.Printf("%d CPUs will be use.\n", n)
		}
	}

	// Пишем профайл для ч/б изображения, профиль цветного не поддерживается «Оперой»
	profile, e := os.OpenFile(oProfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0777)
	if e == nil {
		defer profile.Close()
		profile.WriteString("0:   0  0 0 0 ;\n0:   1  8 0 2 ;\n0:   9 63 0 2 ;\n0:   1 63 2 1 ;\n0:   1 63 1 0;")
	}

	oTmpName := fp.Join(os.TempDir(), "cornet-bolk-"+strconv.Itoa(os.Getpid())+"-")
	oSaved := int64(0)

	// Цикл обработки файлов
	for num, name := range oFileList {
		fmt.Printf("Processing %s ... ", name)

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

		if R := oRadius + 1; R > 1 {
			corner.Copy(im, 0, 0, 0, 0, R, R)
			corner.Copy(im, 0, h-R, 0, R, R, R)
			corner.Copy(im, w-R, 0, R, 0, R, R)
			corner.Copy(im, w-R, h-R, R, R, R, R)
		}

		// Если имена не сохраняем, то заменяем на сгенерированное имя
		if options["keep-name"] == "0" {
			if oLen > 1 {
				name = fmt.Sprintf(oNameMask, num+1)
			} else {
				name = oNameMask
			}
		}

		tmpname := oTmpName + fp.Base(name)
		gray := isgray(im)

		im.Jpeg(tmpname, oQuality)
		im.Destroy()
		im = nil

		// Оптимизация jpeg
		stat, _ := os.Stat(tmpname)
		cmdkeys := []string{"-copy none", "-outfile", name}

		// Для файлов > 10КБ с вероятностью 94% лучшие результаты даёт progressive
		if stat.Size > 10*1024 {
			cmdkeys = append(cmdkeys, "-progressive")
		}

		// Если файл серый, то оптимизируем его как серый
		if gray {
			cmdkeys = append(cmdkeys, "-grayscale", "-scans", oProfile)
		}

		cmdkeys = append(cmdkeys, "-optimize", tmpname)

		// Запускаем jpegtran
		cmd, _ := exec.Run(
			jpegtran.Jpegtran,
			cmdkeys,
			[]string{},
			wd,
			exec.DevNull,
			exec.DevNull,
			exec.DevNull)

		// идея такая — stdout замыкаем на Writer, берём с него данные, следим за EXIF
		// не забыть прочитать EXIF из файла

		cmd.Close()
		outstat, _ := os.Stat(name)

		oSaved += stat.Size - outstat.Size

		os.Remove(tmpname)
		fmt.Println("done")
	}

	if oSaved > 0 {
		fmt.Printf("Saved %d bytes after optimization.\n", oSaved)
	}
}
