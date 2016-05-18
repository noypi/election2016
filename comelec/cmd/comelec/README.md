### running "comelec" tool
```
> comelec example.js
```

#### index region names, example
```
> comelec update_index.js
```

## example example.js
### must initialize comelec data path
```javascript
$path = "/d/dev/res/election2016/www.pilipinaselectionresults2016.com";
```

#### search region names, example
```javascript
$pretty( $search("philippines") );
```

### acquiring comelect data
```javascript
var oRegion = $O('data/regions/OAV.json');
// tries to convert data to string, then print
console.log(JSON.stringify(oRegion));
```

### use $pretty, to pretty print
```javascript
$pretty(oRegion);
```

### supports http://underscorejs.org/
```javascript
console.log("\n------- test underscorejs")
_.each(oRegion.subRegions, function(item){
	console.log("category:", item.categoryName, ", name:", item.name)
})
```

### using $eachls()
```javascript
console.log("\n----- using $eachls")
// list each
var i = 0
$eachls('data/regions', function(item){
	console.log("i=", i, ",", item)
	i++;
	if (3==i) {
		return false // return false to stop
	}
})

console.log("-------- can list folders")
$eachls('', function(item){
	console.log(item)	
})
```

### can also do $list, but can be slow
```javascript
var regions = $list('data/regions')
console.log("regions length=", regions.length)
```


