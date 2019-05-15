$(document).ready(function() {


  

    $(".event").each(function(i, el){

        var e = $(el)

        var path = e.attr("path")
        var time = e.attr("time")

        e.text(new Date(time).toLocaleTimeString())

        e.click(function(){
            $(".event").removeClass("event-selected")
            e.addClass("event-selected")
            playVideo(path)
        });


    })


})
   
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
  