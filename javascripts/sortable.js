$(function() {
	$( "#top200, #qbs, #rbs, #wrs, #tes, #ks, #defs" ).sortable({
		placeholder: "ui-state-highlight",
		cursor: "pointer",
		scroll: true,
		scrollSensitivity: 30
	});
	$( "#top200, #qbs, #rbs, #wrs, #tes, #ks, #defs" ).disableSelection();
});
