$(document).ready(function() {


    $(".date").hide();

    $("#dates").change(function(el) {
        var selected = $(el.target).val()

       selectDate(selected)
    })

    var lastdate =$("#dates option:last").val();
    $("#dates").val(lastdate);
    selectDate(lastdate)
  

    $(".event").each(function(i, el){

        var e = $(el)

        var path = e.attr("path")
        var time = e.attr("time")

        //e.text(new Date(time).toLocaleTimeString())

        e.click(function(){
            $(".event").removeClass("event-selected")
            e.addClass("event-selected")
            playVideo(path)
        });


    })


})

function selectDate(d) {
    $(".date").hide();

    $(".date").each(function(i, el){
        var sd = $(el);
        var dateVal = sd.attr("date")
        if (dateVal == d) {
            sd.show();
        }
    })
}
   
    function eventClickWrapper(path){
        
        return function(){
            console.log(path)
            playVideo(path)
        }
    }

  

    
function playVideo(href){

    var v = $("video").get(0)
    var vs = $("#videosrc")

    v.pause()
    vs.attr("src",href)
    v.load()
    try {
        v.play()
    }
    catch (ex) {

    } finally{}

}
  