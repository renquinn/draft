var auto_refresh = setInterval (
	function () {
		$('#timer').load('timer');
	}, 1000);
