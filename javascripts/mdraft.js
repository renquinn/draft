$(function() {
	$('.mdraft select').change(function() {
		var player = $(this).val();
		$('#input-player').val(player);
	});
});
