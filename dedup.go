package main

import(
  "fmt"
	"io/ioutil"
	"log"
  "os"
  //"crypto/md5"
  "io"
  //"path/filepath"
  //"runtime"
  "github.com/OneOfOne/xxhash"
  "encoding/gob"
  "sort"
  //"html/template"
	"net/http"
  //"sync"
  "strconv"
  "flag"
  "github.com/gorilla/mux"
  "encoding/json"
)

//var wg sync.WaitGroup

type Todo struct {
	Task string
	Done bool
}

type Api_resp struct {
  Success string
  Errors string
}

type Api_list_resp struct {
  List []File_dup_element
  Success bool
}

type Api_delete_file struct {
  Path string
  Hash uint64
}

type Api_delete_req struct {
  Files []Api_delete_file
}

type Api_delete_resp struct {
  Failed_paths []string
  Success bool
}

type Api_list_index_resp struct {
  Index File_dup_element
  Success bool
}

type Api_list_length_resp struct {
  Length int
  Success bool
}

type Fdup_list struct{
  Element []File_dup_element
  Teststr string
}

var m map[string]uint64


type File_info struct{
  Path string
  hash uint64
}


//var s []File_info
var s []File_info



type file_dup_list struct{
  element []File_dup_element
}

type File_dup_element struct{
  Path []string
  Hash uint64
}

var dup_list []File_dup_element


func scanner(root_path string){
  //defer wg.Done()
  if(root_path == "/proc" || root_path == "/dev" || root_path == "/boot"){
    fmt.Println("THIS WILL NOT SCAN PROC, DEV, or BOOT")
    return
  }
  //REWRITE READDIR NO NEED TO OPEN FILES TWICE
  files, err := ioutil.ReadDir(root_path)
  if err != nil {
    log.Fatal(err)
    //return
  }

  for _, file := range files {
    current_file := ""
    if(root_path != "/"){
      current_file = root_path + "/" + file.Name()
    } else {
      current_file = "/" + file.Name()
    }

    checkfileinfo, err := os.Lstat(current_file)
    if err != nil {
      log.Fatal(err)
      //return
    }
    //fmt.Println(checkfileinfo.Mode(), ": File is:", current_file)
    if(checkfileinfo.Mode()&os.ModeSymlink != 0){
      //file_count++;
      continue
    }

    if(file.IsDir()){
      //wg.Add(1)
      //go scanner(current_file)
      scanner(current_file)
    } else{
      //file_count++;
      temp_struct := File_info{current_file, 0}

      s = append(s, temp_struct)
      //m[current_file] = current_file
    }
  }
}

func hasher() {
  for i := 0; i < len(s); i++ {
    f, err := os.Open(s[i].Path)
    if err != nil {
      /*
      if(os.IsPermission(err)){
        f.Close()
        continue
      }
      */
      f.Close()
      log.Fatal(err)
    }

    h := xxhash.New64()
    io.Copy(h, f)
    s[i].hash = h.Sum64()
    f.Close()
  }
}

//taken from a google help pack
//https://groups.google.com/forum/#!topic/golang-nuts/rmKTsGHPjlA
func write_file(file string){
  f, err := os.Create(file)
  if err != nil {
          panic("cant open file")
  }
  defer f.Close()

  enc := gob.NewEncoder(f)
  if err := enc.Encode(dup_list); err != nil {
          panic("cant encode")
  }
}

func read_file(file string) {
  f, err := os.Open(file)
  if err != nil {
          panic("cant open file")
  }
  defer f.Close()

  enc := gob.NewDecoder(f)
  if err := enc.Decode(&dup_list); err != nil {
          panic("cant decode")
  }
}

func check_hash () {
  fmt.Println("I:")
}

func compare_hashes(){

  var temp_dup_files []string
  current_hash := s[0].hash
  temp_dup_files = append(temp_dup_files, s[0].Path)
  for i := 1; i < len(s); i++ {
    if(s[i].hash == current_hash){
      temp_dup_files = append(temp_dup_files, s[i].Path)
      //Add to dup list if this is last loop and temp has more than one file
      if(i + 1 == len(s) && len(temp_dup_files) > 1){
        dup_list = append(dup_list, File_dup_element{temp_dup_files, current_hash})
      }
    } else if (len(temp_dup_files) > 1){
        dup_list = append(dup_list, File_dup_element{temp_dup_files, current_hash})
        current_hash = s[i].hash
        //temp_dup_files = temp_dup_files[:0] //MAY NEED TO FIND A SAFER SLICE CLEAR
        temp_dup_files = nil
        //clear then add first possible dup
        //temp_dup_files = temp_dup_files[:0] //MAY NEED TO FIND A SAFER SLICE CLEAR
        temp_dup_files = append(temp_dup_files, s[i].Path)
        //current_hash := s[i].hash
      } else{
        //temp_dup_files = temp_dup_files[:0] //MAY NEED TO FIND A SAFER SLICE CLEAR
        temp_dup_files = nil
        temp_dup_files = append(temp_dup_files, s[i].Path)
        current_hash = s[i].hash
      }
    }
}

func convert_to_struct() {
  for k, v := range m {
        //fmt.Println("k:", k, "v:", v)
        temp_struct := File_info{k, v}
        s = append(s, temp_struct)
        delete(m, k)
  }
  /*
  for i, v := range s {
    fmt.Println("I:", i)
    fmt.Println("file:", v.path)
    fmt.Println("hash:", v.hash)
  }
  */
}

func convert_to_map() {
  fmt.Println("Hello")
}

func api_list_length(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.Header().Set("Access-Control-Allow-Origin", "*")
  data := Api_list_length_resp {
    len(dup_list),
    true,
  }

  json.NewEncoder(w).Encode(data)
  defer r.Body.Close()
}

func api_list_index(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.Header().Set("Access-Control-Allow-Origin", "*")
  vars := mux.Vars(r)

  success := true
  index, err := strconv.Atoi(vars["index"])
  if err != nil {
    success = false
  }

  if(index > len(dup_list) || index < 0){
    success = false
  }

  var data Api_list_index_resp
  if(success){
    data = Api_list_index_resp {
      dup_list[index],
      true,
    }
  } else {
    data.Success = false
  }

  json.NewEncoder(w).Encode(data)
  defer r.Body.Close()
}

func api_list(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.Header().Set("Access-Control-Allow-Origin", "*")
  vars := mux.Vars(r)

  success := true
  from, err := strconv.Atoi(vars["from"])
  if err != nil {
    success = false
  }

  to, err := strconv.Atoi(vars["to"])
  if err != nil {
    success = false
  }

  if(to <= from || to > len(dup_list) || from < 0){
    success = false
  }

  var data Api_list_resp
  if(success){
    data = Api_list_resp {
      dup_list[from:to],
      true,
    }
  } else {
    data.Success = false
  }


  json.NewEncoder(w).Encode(data)
  defer r.Body.Close()
}

func api_delete(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.Header().Set("Access-Control-Allow-Origin", "*")
  decoder := json.NewDecoder(r.Body)
  var t Api_delete_req
  err := decoder.Decode(&t)
  if err != nil {
      panic(err)
  }

  var failed_paths []string
  for i := 0; i < len(t.Files); i++ {
    for j := 0; j < len(dup_list); j++ {
      for k := 0; k < len(dup_list[j].Path); k++ {
        if(dup_list[j].Path[k] == t.Files[i].Path){
          err := os.Remove(t.Files[i].Path)
          if err != nil {
            failed_paths = append(failed_paths, t.Files[i].Path)
          }
        }
      }
    }
  }

  var data Api_delete_resp
  if(len(failed_paths) > 0){
    data = Api_delete_resp {
      failed_paths,
      false,
    }
  } else {
    data = Api_delete_resp {
      failed_paths,
      true,
    }
  }
  json.NewEncoder(w).Encode(data)

  defer r.Body.Close()
}

func api_index(w http.ResponseWriter, r *http.Request) {
  //fmt.Println("Request:", r.URL.Path[1:])
  //if r.URL.Path[1:] == "" {
    http.ServeFile(w, r, "web/index.html")
  //} else {
    //http.ServeFile(w, r, "web/" + r.URL.Path[1:])
  //}
}

func main() {
  //runtime.GOMAXPROCS(2)


  importPtr := flag.String("import", "", "Import filename")
  exportPtr := flag.String("export", "", "Export filename")
  interfacePtr := flag.String("interface", "", "web or text")

  flag.Parse()

  //args := os.Args
  if(len(flag.Args()) != 1){
    log.Fatal("Must give path.")
  }

  if(*importPtr != ""){
    fmt.Printf("Reading File\n")
    read_file(*importPtr)
    fmt.Printf("Reading complete\n")
  } else {
    fmt.Printf("Scanning Files\n")
    //wg.Add(1)
    //go scanner(flag.Args()[0])
    scanner(flag.Args()[0])
    //wg.Wait()
    for i := 0; i < len(s); i++ {
      if(s[i].Path == "/home/zbblanton/testdup/testdup1"){
        fmt.Printf("Found the file in scanner\n")
      }
    }
    fmt.Printf("Scan Complete\n")
    fmt.Printf("Hashing Files\n")
    hasher()
    fmt.Printf("Hashing Complete\n")
    fmt.Printf("Scanning for Duplicates \n")
    sort.Slice(s, func(i, j int) bool { return s[i].hash < s[j].hash })
    compare_hashes()
    for i := 0; i < len(dup_list); i++ {
      fmt.Println("Hash:", dup_list[i].Hash)
      for j := 0; j < len(dup_list[i].Path); j++ {
        fmt.Println("File", j, ":", dup_list[i].Path[j])
      }
    }

    fmt.Printf("Scan Complete \n")
    if(*exportPtr != ""){
      write_file(*exportPtr)
    }
  }

  if(*interfacePtr == "web"){
    fmt.Printf("Web Service Starting\n")

    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/", api_index)
    //router.HandleFunc("/api", TodoIndex)
    router.HandleFunc("/api/delete", api_delete)
    router.HandleFunc("/api/list/length", api_list_length)
    router.HandleFunc("/api/list/{index}", api_list_index)
    router.HandleFunc("/api/list/{from}/{to}", api_list)

    http.ListenAndServe(":8181", router)
  }



  //read_file()
  //fmt.Println("File", dup_list[1])

/*



  fmt.Printf("Writing to file\n")
  write_file()
  fmt.Printf("Writing Complete\n")

  fmt.Printf("Web Service Starting\n")

  //data := Fdup_list{
  //  dup_list[0:10],
  //  "test1234",
  //}
*/


}
