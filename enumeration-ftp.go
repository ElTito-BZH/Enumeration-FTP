package main

import (
   "fmt"
   "net"
   "io"
   "os"
   "strings"
   "time"
   "bytes"
   "github.com/jlaffaye/ftp"
   )


func telechargement_et_test_upload_FTP ( connexion_FTP *ftp.ServerConn, serveur_FTP string ) {

 liste_dossier_FTP := [] string {"/"}

 for _,dossier := range liste_dossier_FTP {
     list_dir,err := connexion_FTP.List(dossier)
     if err != nil {
        fmt.Printf("Erreur lors de la tentative de lecture du contenu du dossier %s du serveur %s\n",serveur_FTP,dossier)
        fmt.Println(err)
        os.Exit(1)

     }
     for _,element := range list_dir {
         if element.Type.String() == "folder" && dossier != "/" {
            liste_dossier_FTP = append(liste_dossier_FTP,dossier + "/" + element.Name)
         } else if element.Type.String() == "folder" && dossier  == "/" {
            liste_dossier_FTP = append(liste_dossier_FTP, "/" + element.Name)
         }
     }

 }

 string_test_upload := bytes.NewBufferString("Test upload de fichier vers serveur FTP\n")

 for _,dossier := range liste_dossier_FTP {
     if dossier != "/" {
        os.MkdirAll(serveur_FTP + "-FTP/" + dossier,0700)
     }
     // Test Upload dans le dossier FTP courant
     err := connexion_FTP.Stor(dossier + "/test_upload_FTP.txt",string_test_upload)
     if err != nil {
        fmt.Printf("Impossible d'envoyer un fichier dans le dossier %s du serveur %s avec les privilèges actuels\n" ,dossier, serveur_FTP)
     } else {
        fmt.Printf("Envoi de fichier réussi dans le dossier %s du serveur %s\n",dossier,serveur_FTP)
        connexion_FTP.Delete(dossier + "/test_upload_FTP.txt")
     }

     //Test Téléchargement du contenu du dossier FTP courant
     list_dir,err := connexion_FTP.List(dossier)
     if err != nil  && strings.Contains(err.Error(),"550 Access is denied"){
        fmt.Printf("Impossible de lister le contenu du dossier %s du serveur %s avec les privilèges actuels\n",dossier,serveur_FTP)
     } else if err != nil {
       fmt.Printf("Erreur inconnu lors de la tentative de listing du contenu du dossier %s du serveur %s (le message d'erreur est affiché ci-dessous). Veuillez retester manuellement\n",dossier,serveur_FTP)
       fmt.Println(err)

     } else {
       for _,element := range list_dir {
           if element.Type.String() == "file" {
              contenu_fichier_FTP,err := connexion_FTP.Retr(dossier + "/" + element.Name)
              if err != nil && strings.Contains(err.Error(),"229 Entering Extended Passive Mode") {
	         fmt.Printf("Erreur lors du téléchargement du fichier %s/%s du serveur %s. Veuillez le télécharger manuellement\n",dossier,element.Name,serveur_FTP)
              } else if  err != nil {
                 fmt.Printf("Impossible de télécharger le fichier %s/%s du serveur %s avec les privilèges actuels\n",dossier,element.Name,serveur_FTP)
                 fmt.Println(err)
              } else {
                 defer contenu_fichier_FTP.Close()
                 file,_ := os.Create(serveur_FTP + "-FTP/" + dossier + "/" + element.Name)
                 io.Copy(file,contenu_fichier_FTP)
              }
        }
     }

     }


 }

}

func is_FTPserver_alive( serveurFTP string) {
 //Objectif de cette fonction :
  //Vérifier que l'adresse IP / nom DNS saisi soit valide
  //Réaliser un test de connexion sur le port 21 TCP de la machine
  //Afficher un message d'erreur en cas d'échec de la tentative + fermer le programme

 if net.ParseIP(serveurFTP) == nil {

    _,check_error := net.LookupIP(serveurFTP)

    if check_error != nil {
       fmt.Fprintf(os.Stderr, "Erreur lors de la tentative de résolution du nom DNS %s. Veuillez vérifier que ce nom et vos paramètres DNS soient corrects \n",serveurFTP)
       os.Exit(1)
    }
 }

 conn, err := ftp.Dial(serveurFTP + ":21")

 if err != nil {

    fmt.Println ("Le serveur FTP de la machine " + serveurFTP + " est éteint. Veuillez vous assurer d'avoir bien saisi le nom de celui-ci")
    os.Exit(1)
    } else {
      conn.Quit()

    }

}


func main () {


  var username_FTP string = ""
  var password_FTP string = ""
  var serveur_FTP string = ""

  if len(os.Args) == 2 {
    serveur_FTP  = os.Args[1]
    is_FTPserver_alive(serveur_FTP)

  } else if len(os.Args) == 4 {
    serveur_FTP  = os.Args[1]
    username_FTP = os.Args[2]
    password_FTP = os.Args[3]
    is_FTPserver_alive(serveur_FTP)

  } else {
    fmt.Println ("Exécution du script : go run enumeration-smb.go serveur_FTP [ usernameFTP passwordFTP]")
    os.Exit(0)
 }

 err := os.Mkdir(serveur_FTP + "-FTP",0700)
 if err != nil && strings.Contains(err.Error(),"file exists") {

    fmt.Printf("Le dossier %s-FTP existe déjà\n", serveur_FTP)

 } else if err != nil && strings.Contains(err.Error(),"permission denied") {

   fmt.Printf("Impossible de créer le dossier %s-FTP dans le dossier courant. Veuillez vous déplacer dans un dossier où vous avez les droits d'écriture avant de relancer ce script",serveur_FTP)
   os.Exit(1)
 }

 conn, err := ftp.Dial(serveur_FTP + ":21",ftp.DialWithTimeout(5*time.Second))

 if err != nil {

    fmt.Println ("Le serveur FTP de la machine " + serveur_FTP + " est éteint. Veuillez vous assurer d'avoir bien saisi le nom de celui-ci")
    os.Exit(1)
 }

 if username_FTP == "" {

    err = conn.Login("anonymous","anonymous")
    if err != nil {
       fmt.Printf("Impossible de se connecter en FTP à la machine %s sans identifiant\n",serveur_FTP)
       os.Exit(1)
    }

 } else {

   err = conn.Login(username_FTP,password_FTP)
   if err != nil {
      fmt.Printf("L'identifiant %s:%s est incorrect pour la machine %s. Veuillez vérifier la saisie de celui-ci avant de rééxécuter ce script\n", username_FTP,password_FTP,serveur_FTP)
      os.Exit(1)
   }
 }

 telechargement_et_test_upload_FTP(conn,serveur_FTP)

 fmt.Printf("Fin du téléchargement des fichiers du serveur FTP %s, qui sont disponibles dans le dossier %s-FTP \n",serveur_FTP,serveur_FTP)

}
