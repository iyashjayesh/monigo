
(function ($) {
    "use strict";

    /*---------------------------------------------------------------------
    Page Loader
    -----------------------------------------------------------------------*/
    jQuery("#load").fadeOut();
    jQuery("#loading").delay().fadeOut("");

   

   

    /*---------------------------------------------------------------------
    Counter
    -----------------------------------------------------------------------*/
    if (window.counterUp !== undefined) {
        const counterUp = window.counterUp["default"]
        const $counters = $(".counter");
        $counters.each(function (ignore, counter) {
            var waypoint = new Waypoint({
                element: $(this),
                handler: function () {
                    counterUp(counter, {
                        duration: 1000,
                        delay: 10
                    });
                    this.destroy();
                },
                offset: 'bottom-in-view',
            });
        });
    }

    // demo box hover

    $('.js-tilt').tilt({

    });



    /*------------------------
        Back to Top
    --------------------------*/

    jQuery('#back-to-top').fadeOut();
    jQuery(window).on("scroll", function () {
        if (jQuery(this).scrollTop() > 250) {
            jQuery('#back-to-top').fadeIn(1400);
        } else {
            jQuery('#back-to-top').fadeOut(400);
        }
    });
    
    /*----------------
    Scroller
    ---------------------*/

    // scroll body to 0px on click
    jQuery('#top').on('click', function () {
        jQuery('top').tooltip('hide');
        jQuery('body,html').animate({
            scrollTop: 0
        }, 800);
        return false;
    });


    /*---------------------------------------------------------------------
    Wow Animation
    -----------------------------------------------------------------------*/
    var wow = new WOW({
        boxClass: 'wow',
        animateClass: 'animated',
        offset: 0,
        mobile: false,
        live: true
    });
    wow.init();



})(jQuery);
