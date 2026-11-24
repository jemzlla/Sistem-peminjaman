package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	StatusMenunggu = "Menunggu persetujuan"
	StatusDisetujui= "disetujui"
	SstatusDitolak = "ditolak"
)

type Alat struct {
	ID          int `json: "id"`
	Nama     string `json: "nama"`
	Kategori string `json: "kategori"`
	Status   string `json: "status"`
}

type Pengajuan struct {
	ID                  int    `json:"id"`
	IDUser              int    `json:"id_user"`
	IDAlat              int    `json:"id_alat"`
	JenisAlat           string `json:"jenis_alat"`
	NamaAlat            string `json:"nama_alat"`
	TanggalPengajuan    string `json:"tanggal_pengajuan"`
	TanggalPeminjaman   string `json:"tanggal_peminjaman"`
	TanggalPengembalian string `json:"tanggal_pengembalian"`
	KeperluanPeminjaman string `json:"keperluan_peminjaman"`
	StatusPeminjaman    string `json:"status_peminjaman"`
	IDAdmin             int    `json:"id_admin,omitempty"`
} 

var (
	mu sync.Mutex 
	alat = []Alat{
		{1, "Laptop LOQ", "Laptop", "tersedia"},
		{2, "Digicam", "Kamera", "dipinjam"},
		{112, "Mouse AZZAK", "Aksesoris", "tersedia"},
		{887, "Mikrofon condenser", "Audio", "tersedia"}
		{5, "Proyektor Epson", "Proyektor", "dipinjam"}
	}  
	pengajuan []Pengajuan              
)

func writeError(w htttp.ResponseWriter, code int, msg string){
	writreJSON(w, code, map[string]string{"error": msg})
}

func getAlat(w http.ResponseWriter, r *http.Request){
	mu.Lock()
	defer mu.Unlock()
	writeJSON(w, http.StatusOK, alat)
}

func postPengajuan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDUser              int    `json:"id_user"`
		IDAlat              int    `json:"id_alat"`
		TanggalPeminjaman   string `json:"tanggal_peminjaman"`
		TanggalPengembalian string `json:"tanggal_pengembalian"`
		KeperluanPeminjaman string `json:"keperluan_peminjaman"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	if req.IDUser == 0 || req.TanggalPeminjaman == "" || req.TanggalPengembalian == "" || req.KeperluanPeminjaman == "" {
		writeError(w, http.StatusBadRequest, "data tidak lengkap")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	var dipilih *Alat
	for i := range alat {
		if alat[i].ID == req.IDAlat {
			dipilih = &alat[i]
			break
		}
	}
	if dipilih == nil {
		writeError(w, http.StatusNotFound, "alat tidak ditemukan")
		return
	}
	if dipilih.Status != "tersedia" {
		writeError(w, http.StatusConflict, "alat sedang dipinjam")
		return
	}

	data := Pengajuan{
		ID:                  len(pengajuan) + 1,
		IDUser:              req.IDUser,
		IDAlat:              dipilih.ID,
		JenisAlat:           dipilih.Kategori,
		NamaAlat:            dipilih.Nama,
		TanggalPengajuan:    time.Now().Format("2006-01-02"),
		TanggalPeminjaman:   req.TanggalPeminjaman,
		TanggalPengembalian: req.TanggalPengembalian,
		KeperluanPeminjaman: req.KeperluanPeminjaman,
		StatusPeminjaman:    StatusMenunggu,
	}
	pengajuan = append(pengajuan, data)
	writeJSON(w, http.StatusCreated, data)
}

func patchPengajuan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDAdmin     int    `json:"id_admin"`
		IDPengajuan int    `json:"id_pengajuan"`
		Keputusan   string `json:"keputusan"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "format JSON tidak valid")
		return
	}

	if req.IDAdmin == 0 || (req.Keputusan != StatusDisetujui && req.Keputusan != StatusDitolak) {
		writeError(w, http.StatusBadRequest, "id_admin wajib diisi dan keputusan harus 'disetujui' atau 'ditolak'")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	var p *Pengajuan
	for i := range pengajuan {
		if pengajuan[i].ID == req.IDPengajuan {
			p = &pengajuan[i]
			break
		}
	}

	if p == nil {
		writeError(w, http.StatusNotFound, "pengajuan tidak ditemukan")
		return
	}

	if p.StatusPeminjaman != StatusMenunggu {
		writeError(w, http.StatusConflict, "pengajuan ini sudah diputuskan")
		return
	}
	p.StatusPeminjaman = req.Keputusan
	p.IDAdmin = req.IDAdmin

	if req.Keputusan == StatusDisetujui {
		for i := range alat {
			if alat[i].ID == p.IDAlat {
				alat[i].Status = "dipinjam"
			}
		}
	}
	writeJSON(w, http.StatusOK, p)
}

func main (){
	mux := http.NewServeMux()
	mux.HandleFunc("GET /alat", getAlat)
	mux.HandleFunc("POST /pengajuan", postPengajuan)
	mux.HandleFunc("PATCH /pengajuan", patchPengajuan)

	log.Println("Server jalan di port 3000")
	log.Fatal(http.ListenAndServe(":3000", mux))
}