package packageGenerativeSelection

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"encoding/csv"
	"strconv"
	"gonum.org/v1/gonum/stat"
	"sort"
)
type Data struct {
	Domainlen int `json:"domainlen"`
	Seed int64 `json:"seed"`
	Score float64 `json:"score"`
	Payload [] Entry `json:"payload"`
}

type Entry struct {
	TextID int `json:"textid"`
	WordCount  int `json:"# words"`
	Genre string `json:"genre"`
	Year int `json:"year"`
}

type SelectionObj struct {
	selectionArray[] int
	fitnessScore float64
}

type SelectionStatObj struct {
	genresWC map[string] int
	years map[int] int
	yearsWC map[int] int
	genres map[string] int
}

func Gen(seed int64, startyear int64, endyear int64) (* Data){
	fmt.Println(seed)
	rand.Seed(seed)
	domain, domain_len := select_domain(startyear, endyear)
	generation := 2000
	inds_count := 64
	inds := make([]*SelectionObj, 0)
	global_min := 999999.999
	datafilename := "./generationData" + strconv.FormatInt(startyear, 10) + ".csv"
	dfile, err := os.Create(datafilename)
	if err != nil {
		fmt.Println(err)
	}
	defer dfile.Close()
	wd := csv.NewWriter(dfile)
	data_record := []string {"Generation", "Score"}
	wd.Write(data_record)
	for i := 0; i < generation; i++ {
		if i  % 10 == 0 {
			fmt.Println("Generation: ", i, " ", len(inds), " individuals")
		}
		if i == 0 {
			for j := 0; j < inds_count; j++{
				inds = append(inds, new_random_selection(domain_len, seed))
			}
		}
		for j := range inds {
			calc_fitness(domain, inds[j], false)
		}
		inds = sort_inds(inds)
		if global_min > inds[0].fitnessScore {
			global_min = inds[0].fitnessScore
		}

		
    	data_record = []string {}
        data_record = append(data_record, strconv.Itoa(i))
        data_record = append(data_record, strconv.FormatFloat(inds[0].fitnessScore,'f', -1, 64))
        wd.Write(data_record)
    	wd.Flush()

		if i % 10 == 0 {
			fmt.Println(inds[0].fitnessScore, "Global min: ", global_min)
		}
		inds = breed(inds, domain, seed)
		
	}
	for j := range inds {
		calc_fitness(domain, inds[j], false)
	}
	inds = sort_inds(inds)
	calc_fitness(domain, inds[0], true)
	selected := print_values(domain, inds)
	data := new(Data)
	data.Domainlen = len(domain)
	data.Seed = seed
	data.Score = inds[0].fitnessScore
	data.Payload = selected
	filename := "./Selection" + strconv.FormatInt(startyear, 10) + ".csv"
	f, err := os.Create(filename)
    if err != nil {
        fmt.Println(err)
    }
    defer f.Close()
	w := csv.NewWriter(f)
	record := []string{"TextId", "WordCount", "Genre", "Year"}
	w.Write(record)
    for _, obj := range selected {
        var record []string
        record = append(record, strconv.Itoa(obj.TextID))
        record = append(record, strconv.Itoa(obj.WordCount))
		record = append(record, obj.Genre)
		record = append(record, strconv.Itoa(obj.Year))
        w.Write(record)
    }
    w.Flush()
	f_end_data, err := os.OpenFile(strconv.FormatInt(startyear, 10) + "_data_output.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	f_end_data.Write([]byte(strconv.FormatFloat(inds[0].fitnessScore,'f', -1, 64)+"\n"))
	f_end_data.Close()


	return data
}

func print_values(domain []Entry, inds []*SelectionObj)([] Entry){
	var selected [] Entry
	for i := 0; i < len(inds[0].selectionArray); i++ {
		if inds[0].selectionArray[i] == 1 {
			selected = append(selected, domain[i])
		}
	}
	fmt.Println(len(selected))
	return selected
}

func breed(inds []*SelectionObj, domain []Entry, seed int64)([]*SelectionObj) {
	for j := range inds {
		calc_fitness(domain, inds[j], false)
	}
	inds = sort_inds(inds)
	splice_point := rand.Intn(len(inds[0].selectionArray))
	offsprings := make([]*SelectionObj, 0)
	offsprings = append(offsprings, inds[0])
	
	for i := 0; i < (len(inds) / 3); i = i + 1 {
		splice_point = rand.Intn(len(inds[0].selectionArray))
		male := make([]int, len(inds[i].selectionArray))
		female := make([]int, len(inds[i].selectionArray))
		copy(male,inds[i].selectionArray)
		breedNum := rand.Intn(len(inds))
		copy(female,inds[breedNum].selectionArray)
		
		offspring1 := new(SelectionObj)
		offspring2 := new(SelectionObj)

		offspring1.selectionArray = append(male[splice_point:], female[:splice_point]...)
		offspring2.selectionArray = append(female[splice_point:], male[:splice_point]...)

		offsprings = append(offsprings,  inds[i], mass_mutate(offspring1, domain, seed),  mass_mutate(offspring2, domain, seed))
	}
	for j := range offsprings {
		calc_fitness(domain, offsprings[j], false)
	}
	inds = sort_inds(offsprings)
	return offsprings
}

func mass_mutate( obj *SelectionObj, domain []Entry, seed int64)(*SelectionObj) {
	mutationChance := 64
	remutate_max := 3
	stats := calc_fitness(domain, obj, false)
	var genresArray [3]string
	genresArray[0] = "NF"
	genresArray[1] = "MAG"
	genresArray[2] = "FIC"

	for j := 0; j < len(genresArray); j++ {
		words_needed := 2000000 - stats.genresWC[genresArray[j]]
		for i := 0; i < len(obj.selectionArray); i++ {
			mutationRoll := rand.Intn(64*64)
			gene := i
			if words_needed > 0 && obj.selectionArray[gene] == 0 && domain[i].Genre == genresArray[j] && mutationRoll < mutationChance {
				obj.selectionArray[gene] = 1
				words_needed = words_needed - domain[i].WordCount
			} else if words_needed < 0 && obj.selectionArray[gene] == 1 && domain[i].Genre == genresArray[j] && mutationRoll < mutationChance {
				obj.selectionArray[gene] = 0
				words_needed = words_needed + domain[i].WordCount
			}
			if words_needed < 1000 && words_needed > 0 {
				break
			}
			if remutate_max > 0 && i == len(obj.selectionArray) - 1 && (words_needed > 1000 || words_needed < 0) {
				remutate_max--
				i = 0
			}
		}
	}

	return obj
}

func sort_inds(inds []*SelectionObj)([]*SelectionObj) {
	sort.Slice(inds[:], func(i, j int) bool {
		return inds[i].fitnessScore < inds[j].fitnessScore
	  })
	return inds
}

func str_to_int(s string)(int){
	i,_:= strconv.Atoi(s)
	return i
}
func select_domain(year_lower_bound int64, year_upper_bound int64) ([]Entry, int) {
	file, _ := os.Open("/path/to/file.csv")
	var entires [] Entry
	var domain [] Entry
	defer file.Close()

	lines, err := csv.NewReader(file).ReadAll()
    if err != nil {
        panic(err)
    }
	for _, line := range lines {
		entires = append(entires, Entry {
			TextID: str_to_int(line[0]),
			WordCount: str_to_int(line[1]),
			Genre: line[2],
			Year: str_to_int(line[3]),
		})
	}

	for k := range entires {
		if in_bounds(entires[k].Year, year_lower_bound, year_upper_bound) && entires[k].Genre != "NEWS"{
			domain = append(domain, entires[k])
		}
	}
	fmt.Println(domain[0])
	return domain, len(domain)
}

func in_bounds(value int, lower_bound int64, upper_bound int64) (bool){
	if int64(value) >= lower_bound && int64(value) <= upper_bound {
		return true
	} else {
		return false
	}
}

func new_random_selection(len int, seed int64)(*SelectionObj) {
	obj := new(SelectionObj)
	obj.selectionArray = make([]int, len)
	for i := range obj.selectionArray {
		selectRoll := rand.Float64()
		if selectRoll < .20{
			obj.selectionArray[i] = 1
		}	
	}
	return obj
}

func print_selectionObj(obj *SelectionObj){
	fmt.Println(obj.selectionArray)
	fmt.Println(obj.fitnessScore)
}

func calc_fitness(domain []Entry, obj *SelectionObj, print bool)(*SelectionStatObj){
	stat := new(SelectionStatObj)
	stat.years = make(map[int] int)
	stat.yearsWC = make(map[int] int)
	stat.genresWC = make(map[string] int)
	stat.genres = make(map[string] int)
	for i := range obj.selectionArray {
		if obj.selectionArray[i] == 1 {
			stat.years[domain[i].Year]++
			stat.yearsWC[domain[i].Year] = stat.yearsWC[domain[i].Year] + domain[i].WordCount
			stat.genresWC[domain[i].Genre] = stat.genresWC[domain[i].Genre] + domain[i].WordCount
			stat.genres[domain[i].Genre]++
		}
	}
	if print {
		fmt.Println(stat.years)
		fmt.Println(stat.genresWC)
		fmt.Println(stat.genres)
		fmt.Println(stat.yearsWC)
	}
	
	obj.fitnessScore = fitnessScore(stat)
	return stat
}
func fitnessScore(stats *SelectionStatObj)(float64){

	yearsVector := make([]float64, 0, len(stats.years))
	for _, value := range stats.years {
		yearsVector= append(yearsVector, float64(value))
	}
	yearsScore := stat.StdDev(yearsVector,nil)

	yearsWCVector := make([]float64, 0, len(stats.yearsWC))
	for _, value := range stats.yearsWC {
		yearsWCVector = append(yearsWCVector, float64(value))
	}
	yearsWCScore := stat.StdDev(yearsWCVector,nil)

	genresVector := make([]float64, 0, len(stats.genres))
	for _, value := range stats.genres {
		genresVector= append(genresVector, float64(value))
	}
	genresScore :=stat.StdDev(genresVector,nil)

	genresWCVector := make([]float64, 0, len(stats.genresWC))
	var genresWCSum int
	for _, value := range stats.genresWC {
		genresWCVector= append(genresWCVector, float64(value))
		genresWCSum = genresWCSum + value
	}
	genresWCScore :=stat.StdDev(genresWCVector,nil) + math.Abs(float64(6000000 - genresWCSum))
	return yearsScore + genresScore + genresWCScore + yearsWCScore
}
