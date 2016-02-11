$(function () {
    var img = new Image();
    var canvas = document.getElementById('canvas');
    var resultSudoku = $("#result");
    var ctx = canvas.getContext('2d');

    img.onload = function(){
        var MAX_HEIGHT = 500;
        var MAX_WIDTH = 500;

        canvas.width = img.width;
        canvas.height = img.height;
        if (img.height >= img.width && img.height > MAX_HEIGHT) {
            canvas.width /= img.height / MAX_HEIGHT;
            canvas.height = MAX_HEIGHT;
        } else if (img.width >= img.height && img.width > MAX_WIDTH) {
            canvas.height /= img.width / MAX_WIDTH;
            canvas.width = MAX_WIDTH;
        }

        ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
    };

    $('#upload').on("change", function(){
        var files = $(this)[0].files;
        if (files.length === 0) { return; }
        img.src = window.URL.createObjectURL(files[0]);
    });

    if (resultSudoku.length === 1) {
        img.src = resultSudoku.attr('src');
    }
});
