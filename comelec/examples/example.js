// must initialize comelec data path
$path = "/d/dev/res/election2016/www.pilipinaselectionresults2016.com";

// acquiring comelect data
var oRegion = $O('data/regions/OAV.json');

// tries to convert data to string, then print
console.log(JSON.stringify(oRegion));
// use $pretty to pretty print
// $pretty(oRegion);

// supports http://underscorejs.org/
console.log("\n------- test underscorejs")

_.each(oRegion.subRegions, function(item){
	console.log("category:", item.categoryName, ", name:", item.name)
})

console.log("\n----- using $eachls")
// list each
var limit = 3;
$eachls('data/regions', function(item){
	$prn(item)
}, limit)

