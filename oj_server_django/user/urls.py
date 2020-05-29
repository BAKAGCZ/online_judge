from django.urls import path
from . import views
urlpatterns = [
    path('', views.userinfo, name='userinfo'),
    path('ranking/<int:pindex>/', views.ranking, name='ranking'),
    path('login/', views.login, name='login'),
    path('loginhandler/', views.loginhandler, name='loginhandler'),
    path('logout/', views.logout, name='logout'),
    path('register/', views.register, name='register'),
    path('registerhandler/', views.registerhandler, name='registerhandler'),
]
