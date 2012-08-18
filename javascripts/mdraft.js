$(function() {
	$('.mdraft > select').bind('change', function() {
		var player = $(this).val();
		$('#input-player').val(player);
		$('form').submit();
	}
});
