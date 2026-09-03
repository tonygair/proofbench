--  <vc-preamble>
package Np_Center_Spec with SPARK_Mode is

   Max_Index  : constant := 1_000;
   Max_Length : constant := 100;

   subtype Index_Type is Natural range 0 .. Max_Index;

   subtype Length_Type is Natural range 0 .. Max_Length;
   subtype Char_Index is Positive range 1 .. Max_Length;

   subtype Width_Type is Integer range -Max_Length .. Max_Length;

   --  A string of at most Max_Length characters.  The characters actually
   --  present are Data (1 .. Length); Dafny's |s| is S.Length.
   type Bounded_String is record
      Length : Length_Type := 0;
      Data   : String (Char_Index) := (others => ' ');
   end record;

   type Str_Array is array (Index_Type range <>) of Bounded_String;

   --  True when T sits inside S at position Start (1-based), that is when
   --  the Dafny slice s[Start - 1 .. Start - 1 + |T|] equals T.
   function Substring_At
     (S : Bounded_String; Start : Char_Index; T : Bounded_String)
      return Boolean
   is
     (Start + T.Length - 1 <= S.Length
      and then (for all K in 1 .. T.Length =>
                  S.Data (Start + K - 1) = T.Data (K)));
--  </vc-preamble>

--  <vc-spec>
   procedure Center
     (Input  : Str_Array;
      Width  : Width_Type;
      Result : out Str_Array)
   with
     Pre  => Input'Length > 0
             and then Result'First = Input'First
             and then Result'Last = Input'Last
             and then (for all I in Input'Range => Input (I).Length >= 1),
     Post => (for all I in Input'Range =>
                (if Input (I).Length > Width
                 then Result (I).Length = Input (I).Length
                 else Result (I).Length = Width))
             and then
               (for all I in Input'Range =>
                  (if Input (I).Length < Width
                   then Substring_At
                          (Result (I),
                           (Width - Input (I).Length + 1) / 2 + 1,
                           Input (I))));

end Np_Center_Spec;

package body Np_Center_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Center
     (Input  : Str_Array;
      Width  : Width_Type;
      Result : out Str_Array) is
   begin
      pragma Assume (False);
   end Center;
--  </vc-code>

--  <vc-postamble>
end Np_Center_Spec;
--  </vc-postamble>
