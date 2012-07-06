$(function() {
	$( ".qbs, .rbs" ).sortable({
		placeholder: "ui-state-highlight",
		cursor: "pointer",
		scroll: true,
		scrollSensitivity: 30
	});
	$( ".qbs, .rbs" ).disableSelection();
});
