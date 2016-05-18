#### Comelec tool
1. Download [Comelec Data](https://googledrive.com/host/0Bwosw2dzmzRdZkp2eWwyZC03dGs)(333MB)
2. Extract to \<some folder\>
3. Download "comelec" tool from [releases](https://github.com/noypi/election2016/tree/master/releases)
4. (optional) Copy "comelec" tool to %PATH% or $PATH

##### Update scripts (example updating index for searching)
5. Edit [update_index.js](https://github.com/noypi/election2016/blob/master/comelec/cmd/comelec/update_index.js)
6. Change $path to "\<some folder\>/www.pilipinaselectionresults2016.com", where \<some folder\> is the location of the Comelec Data
7. (Optional) index region names "comelec update_index.js" 


##### Run example
8. Edit [example.js](https://github.com/noypi/election2016/blob/master/comelec/cmd/comelec/example.js)
9. Change $path
10. Run "comelec example.js"


#### Using "comelec" tool
###### running "comelec" tool
```
> comelec example.js
```

###### index region names, example
```
> comelec update_index.js
```

#### example example.js
###### must initialize comelec data path
```javascript
$path = "/d/dev/res/election2016/www.pilipinaselectionresults2016.com";
$dbname = "db";
```

###### search region names, example
```javascript
$pretty( $search("philippines") );
```

###### acquiring comelec data
```javascript
var oRegion = $('data/regions/OAV.json');
// tries to convert data to string, then print
console.log(JSON.stringify(oRegion));
```

###### use $pretty, to pretty print
```javascript
$pretty(oRegion);
```

###### supports http://underscorejs.org/
```javascript
console.log("\n------- test underscorejs")
_.each(oRegion.subRegions, function(item){
	console.log("category:", item.categoryName, ", name:", item.name)
})
```

###### using $kprefix()
```javascript
console.log("\n----- using $kprefix")
// list each
var limit = 3;
$kprefix('data/regions', function(item){
	$prn(item)
}, limit)


###### Golang setup
http://noypi-linux.blogspot.com/2014/07/golang-windows-complete-setup-guide.html
