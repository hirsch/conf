conf
====

Package conf parses conf files and offers functions for reading.
###Examples
####Configuration file format:
```
#comment
;comment

[section]
value=key
```

####Loading conf file
```go
data, err := conf.Open("filename.conf")
if err != nil {
	// io or parsing error 
}
```

####Reading data
```go
value, err := data.Read("section", "key")
if err != nil {
	//Section or key does not exist
}
fmt.Println(value)	//Prints value
```
