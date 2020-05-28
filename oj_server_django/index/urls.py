from django.urls import path
from . import views
urlpatterns = [
    path('', views.index, name='index'),
    path('error/', views.error, name='error'),
    path('contact/', views.contact, name='contact'),
]
