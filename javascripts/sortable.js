$(function() {
	$( "#sortable" ).sortable({
		placeholder: "ui-state-highlight",
		cursor: "pointer",
		scroll: true,
		scrollSensitivity: 30
	});
	$( "#sortable" ).disableSelection();
});
