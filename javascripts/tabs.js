$(function() {
	$( "#tabs1" ).tabs({
	});
	var sel = $('.number').html()-1;
	$( "#tabs2" ).tabs({
		selected: sel
	});
});
