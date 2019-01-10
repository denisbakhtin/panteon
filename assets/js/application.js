import xhook from 'xhook';
import $ from 'jquery';
window.jQuery = $;
window.$ = $;
import 'popper.js';
import 'parsleyjs';
import 'bootstrap';
import 'select2';
import Siema from 'siema';

import '../scss/application.scss'

import fontawesome from "@fortawesome/fontawesome";
import fasCog from "@fortawesome/fontawesome-free-solid/faCog";
import fasPencil from "@fortawesome/fontawesome-free-solid/faPencilAlt";
import fasTags from "@fortawesome/fontawesome-free-solid/faTags";
import fasCopy from "@fortawesome/fontawesome-free-solid/faCopy";
import fasUser from "@fortawesome/fontawesome-free-solid/faUser";
import fasPlay from "@fortawesome/fontawesome-free-solid/faPlay";
import fasSignOut from "@fortawesome/fontawesome-free-solid/faSignOutAlt";
import fasEdit from "@fortawesome/fontawesome-free-solid/faEdit";
import fasTimes from "@fortawesome/fontawesome-free-solid/faTimes";
import fasSignIn from "@fortawesome/fontawesome-free-solid/faSignInAlt";
import fasUserPlus from "@fortawesome/fontawesome-free-solid/faUserPlus";
import fasArchive from "@fortawesome/fontawesome-free-solid/faArchive";
import fasHome from "@fortawesome/fontawesome-free-solid/faHome";
import fasEye from "@fortawesome/fontawesome-free-solid/faEye";
import fasCheck from "@fortawesome/fontawesome-free-solid/faCheck";
import fasCalendar from "@fortawesome/fontawesome-free-solid/faCalendarAlt";
import fasChevronUp from "@fortawesome/fontawesome-free-solid/faChevronUp";
import fasChevronLeft from "@fortawesome/fontawesome-free-solid/faChevronLeft";
import fasChevronRight from "@fortawesome/fontawesome-free-solid/faChevronRight";
import fasComments from "@fortawesome/fontawesome-free-solid/faComments";
import fasExclamationTriangle from "@fortawesome/fontawesome-free-solid/faExclamationTriangle";
import fasCart from "@fortawesome/fontawesome-free-solid/faShoppingCart";
import fasSearch from "@fortawesome/fontawesome-free-solid/faSearch";
import fasPhone from "@fortawesome/fontawesome-free-solid/faPhone";
import fasEnvelope from "@fortawesome/fontawesome-free-solid/faEnvelope";
import fasKey from "@fortawesome/fontawesome-free-solid/faKey";
import fasAt from "@fortawesome/fontawesome-free-solid/faAt";
import fasEllipsish from "@fortawesome/fontawesome-free-solid/faEllipsisH";

fontawesome.library.add(fasEllipsish, fasChevronLeft, fasChevronUp, fasCog, fasPencil, fasTags, fasComments, fasExclamationTriangle, fasCopy, fasCheck, fasChevronRight, fasCalendar, fasEye, fasUser, fasPlay, fasSignOut, fasEdit, fasTimes, fasSignIn, fasUserPlus, fasArchive, fasHome, fasCart, fasSearch, fasPhone, fasEnvelope, fasKey, fasAt);

$(document).ready(function () {

    //make dropdown link navigatable
    $('.navbar .dropdown-toggle').click(function () {
        if (!isMobileDevice())
            window.location = $(this).attr('href');
    });

    if (document.querySelector('#ck-content')) {
        //add csrf protection to ckeditor uploads
        xhook.before(function (request) {
            if (!/^(GET|HEAD|OPTIONS|TRACE)$/i.test(request.method)) {
                request.xhr.setRequestHeader("X-CSRF-TOKEN", window.csrf_token);
            }
        });

        ClassicEditor
            .create(document.querySelector('#ck-content'), {
                language: 'ru', //to set different lang include <script src="/public/js/ckeditor/build/translations/{lang}.js"></script> along with core ckeditor script
                ckfinder: {
                    uploadUrl: '/admin/upload'
                },
            })
            .catch(error => {
                console.error(error);
            });
    }

    $('#upload').change(function (e) {
        var formData = new FormData();
        var file = document.getElementById('upload').files[0];
        formData.append('upload', file, file.name);
        var xhr = new XMLHttpRequest();
        xhr.open('POST', '/admin/new_image', true);
        // Set up a handler for when the request finishes.
        xhr.onload = function () {
            if (xhr.status === 200) {
                var img = JSON.parse(this.responseText);
                $('#product-form').append('<input type="hidden" name="imageids" value="' + img.ID + '" id="imageids-' + img.ID + '" />');
                $("#images").append('<div class="img-wrapper" data-id="' + img.ID + '"><img src="' + img.URL + '" class="card-img-top" /><div class="text-center mb-2 mt-auto"><a href="#" class="btn btn-outline-secondary btn-sm remove-btn">Удалить</a></div></div>');
                //append img-wrapper
                setDefaultImage();
            } else {
                alert('Возникла ошибка при загрузке файла!');
            }
        };
        // Send the Data.
        xhr.send(formData);
    });

    $('#product-form').on('click', '.img-wrapper', function () {
        $('.img-wrapper.default').removeClass("default");
        $(this).addClass("default");
        setDefaultImage();
    });
    $('#product-form').on('click', '.remove-btn', function () {
        var id = $(this).closest('.img-wrapper').attr('data-id');
        var xhr = new XMLHttpRequest();
        xhr.open('POST', '/admin/images/' + id + '/delete', true);
        // Set up a handler for when the request finishes.
        xhr.onload = function () {
            if (xhr.status === 201) {
                $('.img-wrapper[data-id="' + id + '"]').remove();
                $('#imageids-' + id).remove();
                setDefaultImage();
            } else {
                alert('Возникла ошибка при удалении файла!');
            }
        };
        // Send the Data.
        xhr.send(null);
        return false;
    });

    $('.products-show .image-previews').on('click', 'a', function () {
        $('.image-wrapper a').attr('data-featherlight', $(this).attr('href'));
        $('.image-wrapper img').attr('src', $(this).attr('href'));
        return false;
    })

    $('.table.table-hover tr').on('click', function () {
        var url = $(this).attr('data-url');
        if (url)
            window.location = url;
    });
    $('.table-hover tr .btn').on('click', function (e) {
        e.stopPropagation();
    });

    //siema 
    if ($('#our-works-carousel').length > 0) {
        const ourWorksSiema = new Siema({
            selector: '#our-works-carousel',
            duration: 200,
            easing: 'ease-out',
            perPage: {
                100: 2,
                576: 3,
                768: 4,
            },
            startIndex: 0,
            draggable: true,
            multipleDrag: true,
            threshold: 20,
            loop: true,
            rtl: false,
            onInit: () => {},
            onChange: () => {},
        });
        document.querySelector('.our-works-prev').addEventListener('click', () => ourWorksSiema.prev());
        document.querySelector('.our-works-next').addEventListener('click', () => ourWorksSiema.next());
        setInterval(() => ourWorksSiema.next(), 4000);
    }

    //product-preview image selector
    $('.product-preview-wrapper .images img').on('click', function (e) {
        e.stopPropagation();
        $(this).closest('.product-preview-wrapper').find('.card-image-top').attr('src', $(this).attr('src'));
    });

});

function setDefaultImage() {
    var definput = $('#product-form #default-image-id');
    var def = $('#product-form .img-wrapper.default');
    var defid = (def.length > 0) ? def.attr("data-id") : 0;
    if (defid == 0) {
        var wrapper = $("#product-form .img-wrapper").first();
        defid = (wrapper.length > 0) ? wrapper.attr("data-id") : 0;
    }
    $("#product-form .img-wrapper[data-id='" + defid + "']").addClass("default");
    definput.val(defid);
}

// Write your Javascript code.
function isMobileDevice() {
    return (typeof window.orientation !== "undefined") || (navigator.userAgent.indexOf('IEMobile') !== -1);
};