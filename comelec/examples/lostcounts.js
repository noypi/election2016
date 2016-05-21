$path = "/d/dev/res/election2016/www.pilipinaselectionresults2016.com";
$dbname = "db";


function getCast(otally) {
	var oret = {};
	oret['complete'] = 0<=otally.stats.transmission.indexOf('100.');
	if (!oret.complete) {
		oret['cast'] = 0;
		return oret;
	}
	
	var rows = otally.stats.regionInfo.rows;
	for(var i=0; i<rows.length; i++) {
		if ('ballots.cast'==rows[i].label) {
			oret['cast'] = parseInt(rows[i].value);
		}
	}
	
	return oret;
}

function getResultsTotal(otally) {
	for (var i=0; i<otally.results; i++){
		
	}
}

$kprefix("tallies/", function(k) {
	var otally = $(k);
	$pretty( otally );
	//$pretty( getCast(otally) );
},1);

var q = $Q(function(){
	$prn("in fn")
	return 'mamay';
})

$prn( "fromjs id=", q.$$id );
q.then(function(p){
	$prn('result is =', p);
});