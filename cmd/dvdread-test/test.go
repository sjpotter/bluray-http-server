package main

/*
#cgo pkg-config: dvdread

#include <stdio.h>
#include <errno.h>
#include <stdlib.h>

#include <dvdread/dvd_reader.h>
#include <dvdread/ifo_types.h>
#include <dvdread/ifo_read.h>
#include <dvdread/dvd_udf.h>
#include <dvdread/nav_read.h>
#include <dvdread/nav_print.h>
*/
import "C"
import (
	"fmt"
	"os"
)

func main() {
	 //titleid, _ := strconv.Atoi(os.Args[2])
	 //chapid, _ := strconv.Atoi(os.Args[3])
	 //angle, _ := strconv.Atoi(os.Args[4])

	 data := C.malloc(1024 * C.DVD_VIDEO_LB_LEN)
	 if data == nil {
	 	fmt.Printf("malloc failed")
	 	os.Exit(1)
	 }
	 defer C.free(data)

	 dvd := C.DVDOpen(C.CString(os.Args[1]))
	 if dvd == nil {
	 	fmt.Printf("Couldn't open dvd %v\n", os.Args[1])
	 	os.Exit(1)
	 }
	 defer C.DVDClose(dvd)

	var stat C.dvd_stat_t
	for i := 0; i < 30; i++ {
		ret := C.DVDFileStat(dvd, C.int(i), C.DVD_READ_TITLE_VOBS, &stat)
		if ret != 0 {
			fmt.Printf("DVDFileStat failed: errno = %v\n", C.perror(nil))
			continue
		}
		fmt.Printf("%v = %+v\n", i, stat)
	}

	os.Exit(0)

	/*
	vmg_file := C.ifoOpen(dvd, 0)
	 if vmg_file == nil {
	 	fmt.Printf( "Can't open VMG info.\n" )
        os.Exit(1)
	 }
	 defer C.ifoClose(vmg_file)
	 tt_srpt := vmg_file.tt_srpt

	 tt_srpt_title := (*[1 << 30]C.title_info_t)(unsafe.Pointer(tt_srpt.title))[:tt_srpt.nr_of_srpts:tt_srpt.nr_of_srpts]

	 fmt.Printf("There are %d titles on this DVD.\n", tt_srpt.nr_of_srpts)

	 if titleid < 0 || titleid >= int(tt_srpt.nr_of_srpts) {
        fmt.Printf( "Invalid title %v.\n", titleid + 1)
        os.Exit(1)
	 }
	 fmt.Printf("There are %v chapters in this title.\n", tt_srpt_title[titleid].nr_of_ptts)

	 */

	 /*

	 if chapid < 0 || chapid >= int(tt_srpt_title[titleid].nr_of_ptts) {
	 	fmt.Printf("Invalid chapter %d\n", chapid + 1);
	 	os.Exit(1)
	 }

	 fmt.Printf("There are %v angles in this title.\n", tt_srpt_title[titleid].nr_of_angles)
	 if angle < 0 || angle >= int(tt_srpt_title[titleid].nr_of_angles) {
        fmt.Printf("Invalid angle %v\n", angle + 1);
        os.Exit(1)
	 }

	 vts_file := C.ifoOpen(dvd, C.int(tt_srpt_title[titleid].title_set_nr))
	 if vts_file == nil {
        fmt.Printf("Can't open the title %v info file.\n", tt_srpt_title[titleid].title_set_nr)
        os.Exit(1)
	 }
	 defer C.ifoClose(vts_file)

	 ttn := tt_srpt_title[titleid].vts_ttn
	 vts_ptt_srpt := vts_file.vts_ptt_srpt

	 vts_ptt_srpt_title := (*[1 << 30]C.ttu_t)(unsafe.Pointer(vts_ptt_srpt.title))[:ttn:ttn]
	 srpt_title := vts_ptt_srpt_title[ttn-1]
	 fmt.Printf("# srpt_title = %v\n", srpt_title.nr_of_ptts)
	 vts_ptt_srpt_title_ptt := (*[1 << 30]C.ptt_info_t)(unsafe.Pointer(srpt_title.ptt))[:vts_ptt_srpt_title[ttn-1].nr_of_ptts:vts_ptt_srpt_title[ttn-1].nr_of_ptts]

	 pgc_id := vts_ptt_srpt_title_ptt[chapid].pgcn
	 pgn := vts_ptt_srpt_title_ptt[chapid].pgn

	 cur_pgc := vts_file.vts_pgcit.pgci_srp[pgc_id-1].pgc
	 start_cell := cur_pgc.program_map[pgn-1]-1

	 title := C.DVDOpenFile(dvd, tt_srpt_title[titleid].title_set_nr, C.DVD_READ_TITLE_VOBS)

	 if title == nil {
	 	fmt.Printf( "Can't open title VOBS (VTS_%02d_1.VOB).\n", tt_srpt_title[titleid].title_set_nr)
        os.Exit(1)
	 }
	 defer C.DVDCloseFile(title)

	 next_cell := start_cell

	 for cur_cell := start_cell; next_cell < cur_pgc.nr_of_cells; {
	 	cur_cell = next_cell

    	if cur_pgc.cell_playback[cur_cell].block_type == C.BLOCK_TYPE_ANGLE_BLOCK {
    		var i C.int

            cur_cell += angle

            for {
                if cur_pgc.cell_playback[cur_cell+i].block_mode == C.BLOCK_MODE_LAST_CELL {
                    next_cell = cur_cell + i + 1
                    break;
                }
                i++
            }
        } else {
            next_cell = cur_cell + 1;
        }

        for cur_pack := cur_pgc.cell_playback[cur_cell].first_sector; cur_pack < cur_pgc.cell_playback[cur_cell].last_sector; {
			var dsi_pack C.dsi_t
			var next_vobu, cur_output_size C.uint

			len := C.DVDReadBlocks(title, (C.int)(cur_pack), 1, data)
			if len != 1 {
				fmt.Printf("Read failed for block %v\n", cur_pack)
				os.Exit(1)
			}
			//assert( is_nav_pack( data ) );

			C.navRead_DSI(&dsi_pack, &(data[C.DSI_START_BYTE]))
			//assert( cur_pack == dsi_pack.dsi_gi.nv_pck_lbn );

			cur_output_size = dsi_pack.dsi_gi.vobu_ea

			if dsi_pack.vobu_sri.next_vobu != C.SRI_END_OF_CELL {
				next_vobu = cur_pack + (dsi_pack.vobu_sri.next_vobu & 0x7fffffff)
			} else {
				next_vobu = cur_pack + cur_output_size + 1
			}

			//assert( cur_output_size < 1024 );
			cur_pack++;

			len = C.DVDReadBlocks(title, (int)(cur_pack), cur_output_size, data)
			if len != (int)(cur_output_size) {
				fmt.Printf("Read failed for %v blocks at %v\n", cur_output_size, cur_pack)
				os.Exit(1)
			}

			//fwrite( data, cur_output_size, C.DVD_VIDEO_LB_LEN, stdout)
			cur_pack = next_vobu;
		}
	}
	  */
}