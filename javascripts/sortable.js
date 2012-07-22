$(function() {

// This might not work since right now there is no way
// to pull the players rank
/*
	$( "#top200, #myteam1" ).sortable({
		cursor: "move",
		scrollSensitivity: 1,
		scrollSpeed: 10,
		distance: 10,
		connectWith: '#myteam1',
		revert: true,
		tolerance: 'pointer',
		remove: function(event, ui) {
			var player = $("#pid", ui.item).text();
			$('#input-player').val(player);
			//var position = get the position somehow
			//$('#input-position').val(position)
			$('form').submit();
		}
	}).disableSelection();
	*/

	$( "#qbs, #myteam2" ).sortable({
		cursor: 'move',
		scrollSensitivity: 1,
		scrollSpeed: 10,
		distance: 10,
		connectWith: '#myteam2',
		revert: true,
		tolerance: 'pointer',
		remove: function(event, ui) {
			var player = $("#pid", ui.item).text();
			$('#input-player').val(player);
			$('#input-position').val("qb");
			$('form').submit();
		}
	}).disableSelection();

	$( "#rbs, #myteam3" ).sortable({
		cursor: "move",
		scrollSensitivity: 1,
		scrollSpeed: 10,
		distance: 10,
		connectWith: '#myteam3',
		revert: true,
		tolerance: 'pointer',
		remove: function(event, ui) {
			var player = $("#pid", ui.item).text();
			$('#input-player').val(player);
			$('#input-position').val("rb");
			$('form').submit();
		}
	}).disableSelection();

	$( "#wrs, #myteam4" ).sortable({
		cursor: "move",
		scrollSensitivity: 1,
		scrollSpeed: 10,
		distance: 10,
		connectWith: '#myteam4',
		revert: true,
		tolerance: 'pointer',
		remove: function(event, ui) {
			var player = $("#pid", ui.item).text();
			$('#input-player').val(player);
			$('#input-position').val("wr");
			$('form').submit();
		}
	}).disableSelection();

	$( "#tes, #myteam5" ).sortable({
		cursor: "move",
		scrollSensitivity: 1,
		scrollSpeed: 10,
		distance: 10,
		connectWith: '#myteam5',
		revert: true,
		tolerance: 'pointer',
		remove: function(event, ui) {
			var player = $("#pid", ui.item).text();
			$('#input-player').val(player);
			$('#input-position').val("te");
			$('form').submit();
		}
	}).disableSelection();

	$( "#ks, #myteam6" ).sortable({
		cursor: "move",
		scrollSensitivity: 1,
		scrollSpeed: 10,
		distance: 10,
		connectWith: '#myteam6',
		revert: true,
		tolerance: 'pointer',
		remove: function(event, ui) {
			var player = $("#pid", ui.item).text();
			$('#input-player').val(player);
			$('#input-position').val("k");
			$('form').submit();
		}
	}).disableSelection();

	$( "#defs, #myteam7" ).sortable({
		cursor: "move",
		scrollSensitivity: 1,
		scrollSpeed: 10,
		distance: 10,
		connectWith: '#myteam7',
		revert: true,
		tolerance: 'pointer',
		remove: function(event, ui) {
			var player = $("#pid", ui.item).text();
			$('#input-player').val(player);
			$('#input-position').val("def");
			$('form').submit();
		}
	}).disableSelection();
});
