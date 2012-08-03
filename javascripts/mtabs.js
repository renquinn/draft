$(function() {
	$('#teams > option').each(function() {
		var hiddenTeam = $(this).val();
		$(hiddenTeam).css("display", "none");
	});
	var team = $('#teamnumber').html();
	$("#tabs2-" + team).css("display", "block");

	$('#teams').bind('change', function() {
		$('#teams > option').each(function() {
			var hiddenTeam = $(this).val();
			$(hiddenTeam).css("display", "none");
		});
		var team = $('#teams select').val();
		$(team).css("display", "block");
	});
});
